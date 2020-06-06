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
	conf.data = make(map[string]string)
	err := parseConfFile(conf, model)
	if err != nil {
		e := ConfError{"conf parse failed"}
		return conf, e
	}

	return conf, nil
}

/**
 * @Description: 加载配置文件
 * @param {type} 
 * @return: 
 */
func parseConfFile(conf *GlassConf, model string) error {
	var sectionSlice []string

	//read file
	fd, err := os.Open(conf.fileName)
	if err != nil {
		fmt.Println("conf file open failed")
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
		if len(line) == 0 {
			continue
		}
		
		str := string(line)
		//处理配置项
		sectionContent, isSection := isSection(str)
		if ! isSection {
			option := strings.Split(str, "=")
			if len(option) != 2 {
				fmt.Println("conf file option format error")
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
		level, sectionName, err := getLevel(sectionContent)
		if err != nil {
			fmt.Println("conf file section format error, illegal secion="+str)
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

	return nil
}

/**
 * @Description: 获取配置段的层级
 * @param {type} 
 * @return: 
 */
func getLevel(confSection string) (int, string, error) {
	var level int
	var sectionName string

	index := strings.LastIndex(confSection, ".")
	//不存在段名
	if index == strings.Count(confSection, "") - 2 {
		e := ConfError{"conf file section format error"}
		return level, sectionName, e;
	}
	if index == -1 {
		sectionName = confSection
		level = 0
	} else {
		sectionName = confSection[index+1:]
		level = index + 1
	}

	return level, sectionName, nil
}

/**
 * @Description: 判断是否是一个配置段名
 * @param {type} 
 * @return: 
 */
func isSection(confStr string) (string, bool) {
	sectionRegex := regexp.MustCompile(`^\[([.]*[a-zA-Z0-9_-]+)\]$`)
	params := sectionRegex.FindStringSubmatch(confStr)
	if len(params) != 2 {
		return "", false;
	}

	return params[1], true
}

func (err ConfError) Error() string {
    strFormat := `
    glass_conf error, msg=%s
`
    return fmt.Sprintf(strFormat, err.errMsg)
}