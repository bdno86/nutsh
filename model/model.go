package model

import (
	"os"
	"fmt"
	"io/ioutil"
	"strconv"
	"regexp"
	"github.com/blinry/goyaml"
	"morr.cc/nutsh.git/parser"
	"morr.cc/nutsh.git/cli"
	//"sort"
)

type Tutorial struct {
	Name    string
	Target  string
	Version int
	Basedir string
	Lessons map[string]*Lesson
	Common *parser.Node
}

type Lesson struct {
	Root *parser.Node
	Done bool
}

func NameToNumber(name string) int {
	number_string := regexp.MustCompile(`\d+`).FindString(name)
	number, err := strconv.Atoi(number_string)
	if err != nil {
		return -1
	}
	return number
}

func Init(dir string, low int, high int) Tutorial {
	info, _ := ioutil.ReadFile(dir + "/info.yaml")
	var tut Tutorial
	goyaml.Unmarshal(info, &tut)
	tut.Basedir = dir
	tut.Lessons = make(map[string]*Lesson)

	files, _ := ioutil.ReadDir(dir)
	for _, file := range files {
		if len(file.Name()) >= 7 && file.Name()[len(file.Name())-6:len(file.Name())] == ".nutsh" {

			number := NameToNumber(file.Name())
			if (number < low || number > high) && !(file.Name() == "common.nutsh") {
				continue
			}

			content, _ := ioutil.ReadFile(dir + "/" + file.Name())
			rootnode := parser.Parse(string(content))
			if file.Name() == "common.nutsh" {
				tut.Common = rootnode
			} else {
				tut.Lessons[file.Name()[0:len(file.Name())-6]] = &Lesson{rootnode, false}
			}
		}
	}

	var done_lessons []string

	s, err := ioutil.ReadFile(dir+"/progress.yaml")
	if err == nil {
		goyaml.Unmarshal(s, &done_lessons)
	}
	for _, l := range done_lessons {
		l, ok := tut.Lessons[l]
		if ok {
			l.Done = true
		}
	}

	return tut
}

func (t Tutorial) SelectLesson(auto bool) (string, bool) {
	lessons := make([]string, len(t.Lessons))
	for name, _ := range t.Lessons {
		if NameToNumber(name) >= 0 {
			lessons[NameToNumber(name)] = name
		}
	}

	if auto {
		for _, name := range lessons {
			l := t.Lessons[name]
			if ! l.Done {
				return name, true
			}
		}
	}

	fmt.Printf("\n[34m== %s ==[0m\n\n", t.Name)
	for i, name := range lessons {
		l := t.Lessons[name]
		if l.Done {
			fmt.Print("[32m")
		}
		fmt.Printf("%d) ", i+1)
		fmt.Print(l.Name())
		if l.Done {
			fmt.Print("[0m")
		}
		fmt.Println()
	}
	fmt.Println("\n0) [Beenden]")

	sel := 0
tryagain:
	fmt.Print("\nBitte wählen Sie eine Lektion: ")

	input := cli.GetInput()
	buf := make([]rune, 0)
	for {
		r := <-input
		if r != '\u000a' {
			buf = append(buf, r)
		} else {
			break
		}
	}
	sel, err := strconv.Atoi(string(buf))
	if err != nil {
		goto tryagain
	}

	if sel < 0 || sel > len(lessons) {
		goto tryagain
	}

	if sel == 0 {
		return "", false
	}

	return lessons[sel-1], true
}

func (t Tutorial) SaveProgress() {
	done_lessons := make([]string, 0)
	for name, l := range t.Lessons {
		if l.Done {
			done_lessons = append(done_lessons, name)
		}
	}
	s, _ := goyaml.Marshal(done_lessons)
	f, _ := os.Create(t.Basedir+"/progress.yaml")
	f.Write(s)
	f.Close()
}

func (l Lesson) Name() string {
	return parser.GetName(l.Root)
}
