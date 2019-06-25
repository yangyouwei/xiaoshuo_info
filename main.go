package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
)
//并发
var concurrenc int
var filenamech = make(chan string,10)

//book信息
type bookinfo struct {
	Bookname string
	Author string
	Size int
	Booksummary string
	Chapternumber int
	Bookmenu []string
	Storeopath string
}

func main()  {
	pathname,err := filepath.Abs("D:\\a")
	if err != nil  {
		fmt.Println("path error")
		return
	}
	concurrenc = 5
	wg := sync.WaitGroup{} //控制主程序等待，以便goroutines运行完
	wg.Add(concurrenc+1)
	go func(wg *sync.WaitGroup,filenamech chan string) {
		GetAllFile(pathname,filenamech)
		close(filenamech)   //关闭通道，以便读取通道的程序知道通道已经关闭。
		wg.Done()	//一定在函数的内部的最后一行运行。否则可能函数没有执行完毕。
	}(&wg,filenamech)
	for i := 0;i < concurrenc; i++  {
		go func(wg *sync.WaitGroup,filenamech chan string) {
			for {
				filename,isclose := <- filenamech
				if !isclose {	//判断通道是否关闭，关闭则退出循环
					break
				}
				dosomewrork(filename)
			}
			wg.Done()
		}(&wg,filenamech)
	}
	wg.Wait()
	fmt.Println("finish")
}


func GetAllFile(pathname string,fn_ch chan string) {
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		fmt.Println("read dir fail:", err)
	}
	for _, fi := range rd {
		if fi.IsDir() {
			fullDir := pathname + "/" + fi.Name()
			GetAllFile(fullDir,fn_ch)
			if err != nil {
				fmt.Println("read dir fail:", err)
			}
		} else {
			fullName := pathname + "/" + fi.Name()
			fn_ch <- fullName
		}
	}
}

func dosomewrork(fp string)  {
	a := bookinfo{}
	a.getbookinfo(fp)
	a.printinfo()
}

func (b *bookinfo)getbookinfo(fp string)  {
	b.Size = 1024
	b.Author = "马克吐温"
	b.Bookmenu = []string{"第一章","第二章","第三章","第四章","第五章","第六章","第七章"}
	b.Booksummary = "美国作者"
	b.Chapternumber = 100
	b.Bookname = "鲁滨逊历险记"
	b.Storeopath = "/xxx.txt"
}

func (b *bookinfo)printinfo()  {
	fmt.Println(b.Bookname)
	fmt.Println(b.Chapternumber)
	fmt.Println(b.Booksummary)
	fmt.Println(b.Author)
	fmt.Println(b.Size)
}

func (b *bookinfo)wirteinfo()  {

}