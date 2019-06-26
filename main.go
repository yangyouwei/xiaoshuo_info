package main

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)
//并发
var concurrenc int
var filenamech = make(chan string,10)
var bookstorepath = "/data/zip/books/"


//book信息
type Bookinfo struct {
	Bookid	int64	`db:"id"`
	Bookname string	`db:"bookname"`
	Boookauthor string `db:"boookauthor"`
	Bookcahtpernum	int `db:"bookcahtpernum"`
	Bookcomment	string `db:"bookcomment"`
	Bookstorepath string `db:"bookstorepath"`
}


func main()  {
	pathname,err := filepath.Abs("/data/zip/unrarfull")
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
				a := filepath.Base(filename)
				fmt.Println(a)
				dosomewrork(filename)
			}
			wg.Done()
		}(&wg,filenamech)
	}
	wg.Wait()
	fmt.Println("finish")
}

//获取文件名
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

//文档处理函数
func dosomewrork(fp string)  {
	b := Bookinfo{}
	b.getinfo(fp)
	b.insert()
}

func (b *Bookinfo)getinfo(fp string)  {
	//bookname
	bn := strings.Split(filepath.Base(fp),".")
	bookname := bn[1]
	b.Bookname = bookname
	//作者
	b.Boookauthor = getbookauthor(fp)
	//章节数
	b.Bookcahtpernum = 0
	//摘要
	b.Bookcomment = ""
}

//book信息写入数据库
func (b *Bookinfo)insert()  {
	db, err := sql.Open("mysql", "root:gaopeng@tcp(192.168.2.250:3306)/books")
	check(err)

	stmt, err := db.Prepare(`INSERT bookinfo ( bookname, boookauthor, bookcahtpernum,bookcomment,bookstorepath) VALUES (?,?,?,?,?)`)
	check(err)

	res, err := stmt.Exec(b.Bookname,b.Boookauthor,b.Bookcahtpernum,b.Bookcomment,b.Bookstorepath)
	check(err)

	id, err := res.LastInsertId()  //必须是自增id的才可以正确返回。
	check(err)
	idstr := fmt.Sprintf("%v",id)
	err = os.Mkdir(bookstorepath + idstr, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	stmt, err = db.Prepare("UPDATE bookinfo set bookstorepath=? WHERE id=?")
	check(err)

	res, err = stmt.Exec(bookstorepath + idstr,id)
	check(err)

	_, err = res.RowsAffected()  //返回值是1 修改成功，如果是零则是不成功（修改值和原来的值一样，mysql实际没有修改，或者没有这条记录）
	check(err)

	fmt.Println(idstr)
	stmt.Close()
}

func check(err error) {
	if err != nil{
		fmt.Println(err)
		panic(err)
	}
}

func getbookauthor(fp string) string {
	author := ""
	fi, err := os.Open(fp)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return ""
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	for i:=0;i<20;i++ {		//读取前20行
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		res := getname(string(a))
		if res != "" {
			author = res
			break
		}
	}
	return author
}

//正则表达式
func getname(s string) string {
	isok , err := regexp.Match("作者：",[]byte(s))
	if err != nil {
		fmt.Println(err)
	}
	if isok {
		reg := regexp.MustCompile(".*(作者：)(.*)")   //分组，第一个分组是全部匹配的结果，第二个是括号里的。
		result := reg.FindAllStringSubmatch(s,-1)  //使用for循环然后取切片的下标，或者使用result1[0][1]直接取出
		a := result[0][2]
		fmt.Println(a)
		return a
	}

	isok , err = regexp.Match("著",[]byte(s))
	if err != nil {
		fmt.Println(err)
	}
	if isok {
		reg := regexp.MustCompile("(.*)(著)(\\s*)")   //分组，第一个分组是全部匹配的结果，第二个是括号里的。
		result := reg.FindAllStringSubmatch(s,-1)  //使用for循环然后取切片的下标，或者使用result1[0][1]直接取出
		a := result[0][0]
		fmt.Println(a)
		return a
	}
	return ""
}
