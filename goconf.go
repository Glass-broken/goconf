package goconf

import (
	"fmt"
	"sync"
	"os"
	"io"
	"bufio"
	"strings"
	"regexp"
)


const (
	DEFAULT_SECTION = "DEFAULT"
)

type GlassConf struct {
	lock 		sync.RWMutex
	fileName 	string
	data		map[string]string

	blockMode	bool
}
type ConfError struct {
    errMsg string
}


var allowModel = map[string]int{"default":1}

func LoadConf(fileName string, model string) (*GlassConf, error) {
	isExist := isExist(fileName)
	if false == isExist {
		error := ConfError{"conf file not exists"}
		return nil, error
	}
	conf := new(GlassConf)
	conf.fileName = fileName
	err := parseConfFile(conf, model)
	if err != nil {
		e := ConfError{"conf parse failed"}
		return conf, e
	}

	return conf, nil
}


func parseConfFile(conf *GlassConf, model string) error {
	var sectionSlice []string

	//read file
	fd, err := os.Open(conf.fileName)
	if err != nil {
		e := ConfError{"conf file open failed"}
		return e
	}
	defer fd.Close()
	reader := bufio.NewReader(fd)
	for {
		line, _, isEof := reader.ReadLine()
		if isEof == io.EOF {
			break
		}
		str := string(line)
		//处理配置项
		if ! isSection(str) {
			option := strings.Split(str, ":")
			if len(option) != 2 {
				e := ConfError{"conf file option format error"}
				return e
			}
			prefix := strings.Join(sectionSlice, "/")
			key := string(option[0])
			value := string(option[1])
			prefix += "/" + key
			conf.data[prefix] = value
			continue
		}
		//处理配置段
		level, sectionName, err := getLevel(str)
		if err != nil {
			e := ConfError{"conf file section format error, illegal secion="+str}
			return e
		}
		sectionLen := len(sectionSlice)
		if level == sectionLen {
			sectionSlice = append(sectionSlice, sectionName)
		} else if level < sectionLen {
			//左闭右开区间
			sectionSlice = sectionSlice[:level]
			sectionSlice = append(sectionSlice, sectionName)
		}
	}
}

func getLevel(confStr string) (int, string, error) {
	var level int
	var sectionName string

	sectionRegex := regexp.MustCompile(`^\[([.]*[a-zA-Z0-9_]+)\]$`)
	params := sectionRegex.FindStringSubmatch(confStr)
	if len(params) != 2 {
		e := ConfError{"conf file section format error"}
		return level, sectionName, e;
	}
	index := strings.LastIndex(params[1], ".")
	//不存在段名
	if index == strings.Count(params[1], "") - 2 {
		e := ConfError{"conf file section format error"}
		return level, sectionName, e;
	}
	if index == -1 {
		sectionName = params[1]
		level = 0
	} else {
		sectionName = params[1][index+1:]
		level = index + 1
	}

	return level, sectionName, nil
}

func isSection(confStr string) bool {
	str := strings.TrimSpace(confStr)
	len := strings.Count(str, "") - 1
	if str[0] == '[' && str[len - 1] == ']' {
		return true
	}

	return false
}

func (err ConfError) Error() string {
    strFormat := `
    glass_conf error, msg=%s
`
    return fmt.Sprintf(strFormat, err.errMsg)
}