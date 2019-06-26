package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
)

func main()  {
	//获取文件大小
	//a, err := os.Lstat("")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(a.Size())
	fp := "D:\\aaa.txt"
	getbookauthor(fp)
}

func getbookauthor(fp string){
	fi, err := os.Open(fp)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	for {		//读取前20行
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		//fmt.Println(string(a))
		getname(string(a))
	}
}

func getname(s string) {
	isok , err := regexp.Match("^(\\s*)(第)(.*)(章)(.*)",[]byte(s))
	if err != nil {
		fmt.Println(err)
	}

	if isok {
		reg := regexp.MustCompile("^(\\s*)(第)(.*)(章)(.*)$")    //分组，第一个分组是全部匹配的结果，第二个是括号里的。
		result := reg.FindAllStringSubmatch(s,-1)  //使用for循环然后取切片的下标，或者使用result1[0][1]直接取出
		a := result[0][0]
		fmt.Println(a)
	}
}
