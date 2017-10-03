package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/cloudflare/ahocorasick"
	"github.com/fatih/color"
	"github.com/remeh/sizedwaitgroup"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//a matcher contains the regex for comment tags we are looking for
type matcher struct {
	Regex string
	Label string
	Tag   string
}

//Result holds data for all matches found in a file
type Result struct {
	Filename string
	Matches  []match
}

//a match contains data for each comment tag found in a file
type match struct {
	lineNumber int
	tag        string
	label      string
	author     string
	message    string
}

var (
	version         string
	skipHidden      *bool
	ignoreDirs      *[]string
	ignoreExts      *[]string
	lineLengthLimit *int

	outputMux sync.Mutex

	matchers   []matcher
	tagMatcher *ahocorasick.Matcher
)

func main() {

	app := kingpin.New("fixme", "Searches for comment tags in code")
	filePath := app.Arg("file", "the file or directory to scan").Required().String()
	skipHidden = app.Flag("skip-hidden", "skip hidden folders (default=true)").Default("true").PlaceHolder("true").Bool()
	ignoreDirs = app.Flag("ignore-dir", "pattern of directories to ignore").Short('i').PlaceHolder("vendor").Strings()
	ignoreExts = app.Flag("ignore-exts", "pattern of file extensions to ignore").PlaceHolder(".txt").Strings()
	lineLengthLimit = app.Flag("line-length-limit", "number of max characters in a line").Default("1000").Int()
	logLvl := app.Flag("log-level", "log level (debug|info|error)").Short('l').Default("error").Enum("debug", "info", "error")

	app.Version(version).VersionFlag.Short('V')
	app.HelpFlag.Short('h')
	app.UsageTemplate(kingpin.SeparateOptionalFlagsUsageTemplate)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	log.SetHandler(cli.New(os.Stdout))
	log.SetLevelFromString(*logLvl)

	//set up the regex values
	matchers = initMatchers()
	tagMatcher = ahocorasick.NewStringMatcher([]string{"NOTE", "OPTIMIZE", "TODO", "HACK", "XXX", "FIXME", "BUG"})

	//get the files from the path the user specified
	cleanPath, err := filepath.Abs(*filePath)
	if err != nil {
		log.WithError(err).Fatal("Error identifying files")
	}

	fileList := getFiles(cleanPath)

	wg := sizedwaitgroup.New(runtime.NumCPU())
	for _, file := range fileList {
		wg.Add()
		go func(file string) {
			defer wg.Done()

			//scan for comment tags within the file
			result, err := processFile(file)
			if err != nil {
				log.WithError(err).Errorf("Error processing %s", file)
			}

			//print results from the file
			printMatches(result)
		}(file)
	}
	wg.Wait()

}

func printMatches(result Result) {
	//acquire a lock to ensure our output is stable
	outputMux.Lock()
	defer outputMux.Unlock()

	numOfMatches := len(result.Matches)
	if numOfMatches == 0 {
		return
	}

	color.Set(color.FgHiWhite, color.Bold)
	fmt.Printf("• %s ", result.Filename)
	color.Unset()

	color.Set(color.Faint)
	if numOfMatches == 1 {
		fmt.Printf("[%d message]:\n", numOfMatches)
	} else {
		fmt.Printf("[%d messages]:\n", numOfMatches)
	}
	color.Unset()

	for _, m := range result.Matches {
		color.Set(color.Faint)
		fmt.Printf(" [Line %d]\t", m.lineNumber)
		color.Unset()

		switch m.tag {
		case "NOTE":
			color.Set(color.Bold, color.FgHiGreen)
			if len(m.author) > 0 {
				fmt.Printf(" %s from %s:", m.label, m.author)
			} else {
				fmt.Printf(" %s:", m.label)
			}
			color.Unset()
			color.Set(color.FgGreen)
			fmt.Printf(" %s\n", m.message)
			color.Unset()
		case "OPTIMIZE":
			color.Set(color.Bold, color.FgHiBlue)
			if len(m.author) > 0 {
				fmt.Printf(" %s from %s:", m.label, m.author)
			} else {
				fmt.Printf(" %s:", m.label)
			}
			color.Unset()
			color.Set(color.FgBlue)
			fmt.Printf(" %s\n", m.message)
			color.Unset()
		case "TODO":
			color.Set(color.Bold, color.FgHiMagenta)
			if len(m.author) > 0 {
				fmt.Printf(" %s from %s:", m.label, m.author)
			} else {
				fmt.Printf(" %s:", m.label)
			}
			color.Unset()
			color.Set(color.FgHiMagenta)
			fmt.Printf(" %s\n", m.message)
			color.Unset()
		case "HACK":
			color.Set(color.Bold, color.FgHiYellow)
			if len(m.author) > 0 {
				fmt.Printf(" %s from %s:", m.label, m.author)
			} else {
				fmt.Printf(" %s:", m.label)
			}
			color.Unset()
			color.Set(color.FgYellow)
			fmt.Printf(" %s\n", m.message)
			color.Unset()
		case "XXX":
			color.Set(color.Bold, color.FgHiCyan)
			if len(m.author) > 0 {
				fmt.Printf(" %s from %s:", m.label, m.author)
			} else {
				fmt.Printf(" %s:", m.label)
			}
			color.Unset()
			color.Set(color.FgCyan)
			fmt.Printf(" %s\n", m.message)
			color.Unset()
		case "FIXME":
			color.Set(color.Bold, color.FgHiRed)
			if len(m.author) > 0 {
				fmt.Printf(" %s from %s:", m.label, m.author)
			} else {
				fmt.Printf(" %s:", m.label)
			}
			color.Unset()
			color.Set(color.FgRed)
			fmt.Printf(" %s\n", m.message)
			color.Unset()
		case "BUG":
			fmt.Print("  ")
			color.Set(color.Bold, color.FgWhite, color.BgRed)
			if len(m.author) > 0 {
				fmt.Printf("%s from %s:", m.label, m.author)
			} else {
				fmt.Printf("%s:", m.label)
			}
			color.Unset()
			fmt.Print(" ")
			color.Set(color.FgRed)
			fmt.Printf("%s\n", m.message)
			color.Unset()
			fmt.Println()
		}

	}
	fmt.Println()
}
func processFile(file string) (Result, error) {

	var result Result
	result.Filename = file

	log.Debug("Processing " + file)
	f, err := os.Open(file)
	if err != nil {
		return result, err
	}

	scanner := bufio.NewScanner(f)
	lineNumber := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++

		//skip if the line is too long
		if len(line) > *lineLengthLimit {
			continue
		}

		//check the line with the MPM before running against regular expressions
		hits := tagMatcher.Match([]byte(line))
		if len(hits) <= 0 {
			continue
		}

		//check the line against the regexes
		for _, m := range matchers {
			var re = regexp.MustCompile(m.Regex)

			if re.MatchString(line) {
				//skip tags with an empty message
				if len(re.FindStringSubmatch(line)[2]) <= 0 {
					continue
				}
				result.Matches = append(result.Matches, match{lineNumber, m.Tag, m.Label, re.FindStringSubmatch(line)[1], re.FindStringSubmatch(line)[2]})
			}
		}
	}

	return result, nil
}

// getFiles returns a list of the files to be processed
func getFiles(filePath string) []string {

	fileList := []string{}
	err := filepath.Walk(filePath, func(path string, f os.FileInfo, err error) error {

		//catch any errors while walking the filepath
		if err != nil {
			return err
		}

		//skip hidden directories if the user requests it
		if *skipHidden {
			if f.IsDir() && strings.HasPrefix(f.Name(), ".") {
				return filepath.SkipDir
			}
		}

		//skip any directories the user wants to ignore
		if len(*ignoreDirs) > 0 {
			for _, dir := range *ignoreDirs {
				if f.IsDir() && strings.HasPrefix(f.Name(), dir) {
					return filepath.SkipDir
				}
			}
		}

		//skip any files with an extension the user wants to ignore
		if len(*ignoreExts) > 0 {
			for _, ext := range *ignoreExts {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ext) {
					return filepath.SkipDir
				}
			}
		}

		if !f.IsDir() {
			fileList = append(fileList, path)
			log.Debug("found file: " + path)
		}

		return nil
	})

	if err != nil {
		log.WithError(err).Fatal("failed getting file names")
	}

	return fileList
}

func initMatchers() []matcher {
	return []matcher{
		{
			Regex: `(?i)(?:[\/\/][\/\*]|#)\s*NOTE\b\s*(?:\(([^:]*)\))*\s*:?\s*(.*)`,
			Label: ` ✐ NOTE`,
			Tag:   "NOTE",
		},
		{
			Regex: `(?i)(?:[\/\/][\/\*]|#)\s*OPTIMIZE\b\s*(?:\(([^:]*)\))*\s*:?\s*(.*)`,
			Label: ` ↻ OPTIMIZE`,
			Tag:   "OPTIMIZE",
		},
		{
			Regex: `(?i)(?:[\/\/][\/\*]|#)\s*TODO\b\s*(?:\(([^:]*)\))*\s*:?\s*(.*)`,
			Label: ` ✓ TODO`,
			Tag:   "TODO",
		},
		{
			Regex: `(?i)(?:[\/\/][\/\*]|#)\s*HACK\b\s*(?:\(([^:]*)\))*\s*:?\s*(.*)`,
			Label: ` ✄ HACK`,
			Tag:   "HACK",
		},
		{
			Regex: `(?i)(?:[\/\/][\/\*]|#)\s*XXX\b\s*(?:\(([^:]*)\))*\s*:?\s*(.*)`,
			Label: ` ✗ XXX`,
			Tag:   "XXX",
		},
		{
			Regex: `(?i)(?:[\/\/][\/\*]|#)\s*FIXME\b\s*(?:\(([^:]*)\))*\s*:?\s*(.*)`,
			Label: ` ☠ FIXME`,
			Tag:   "FIXME",
		},
		{
			Regex: `(?i)(?:[\/\/][\/\*]|#)\s*BUG\b\s*(?:\(([^:]*)\))*\s*:?\s*(.*)`,
			Label: `☢ BUG`,
			Tag:   "BUG",
		},
	}
}
