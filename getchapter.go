package main

import (
	"bufio"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

//book信息
type Bookinfo struct {
	Bookname        string `db:"bookname"`
	Boookauthor     string `db:"boookauthor"`
	Bookcahtpernum  int    `db:"bookcahtpernum"`
	Bookcomment     string `db:"bookcomment"`
	Bookstorepath   string `db:"bookstorepath"`
}
//章节信息
type chapterinfo struct {
	Bookid	int `db:"bookid"`
	Chapterid int `db:"chapterid"`
	Chaptername string `db:"chaptername"`
	Chapter string `db:"chapter"`
	Storepath string `db:"storepath"`
	Chapterlines int64 `db:"chapterlines"`
}

//文件名chan
var filenamech = make(chan string, 100)

//mysql连接池
var dbcon *sql.DB

//初始化mysql连接池
func init()  {
	db, err := sql.Open("mysql", "root:gaopeng@tcp(192.168.2.250:3306)/books")
	dbcon = db
	check(err)
}


func main() {
	//获取绝对路径
	pathname, err := filepath.Abs("/data/zip/unrarfull")
	if err != nil {
		fmt.Println("path error")
		return
	}
	//并发数量
	concurrenc := 5

	wg := sync.WaitGroup{} //控制主程序等待，以便goroutines运行完
	wg.Add(concurrenc + 1)
	//开启一个go 程去取文件名（全路径）
	go func(wg *sync.WaitGroup, filenamech chan string) {
		GetAllFile(pathname, filenamech)
		close(filenamech) //关闭通道，以便读取通道的程序知道通道已经关闭。
		wg.Done()         //一定在函数的内部的最后一行运行。否则可能函数没有执行完毕。
	}(&wg, filenamech)
	//for循环开启并发的go程
	for i := 0; i < concurrenc; i++ {
		go func(wg *sync.WaitGroup, filenamech chan string) {
			for {	//for不断重chan中取文件路径
				filename, isclose := <-filenamech   //判断chan是否关闭，关闭了就退出循环不在取文件名结束程序
				if !isclose { //判断通道是否关闭，关闭则退出循环
					break
				}
				//真正干事的
				dosomewrork(filename)
			}
			wg.Done()
		}(&wg, filenamech)
	}
	wg.Wait()
}


func dosomewrork(fp string)  {
	//读取文本
	fi, err := os.Open(fp)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	var lines int64  //文本的行数
	var chapternum int	//找到章节数，作为章节id
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		lines++
		//获取一个chapter结构实例，并存入数据库
		res := getchapterinfo(string(a),fp,lines)
		if res.Chaptername != "" && res.Bookid != 0 {
			chapternum++
			res.Chapterid = chapternum
			res.writetodb()
		}
		continue
	}
}

func getchapterinfo(s string,fp string,lines int64) *chapterinfo {
	c := chapterinfo{}
	c.Bookid, c.Storepath = getbookid(fp)
	c.Chaptername = getchaptername(s)
	c.Chapterlines = lines
	return &c
}

func getchaptername(s string) string {
	isok , err := regexp.Match("^(\\s*)(第)(.*)(章)(.*)|^(\\s*)(\\d)(\\s*)",[]byte(s))
	if err != nil {
		fmt.Println(err)
	}

	if isok {
		reg := regexp.MustCompile("^(\\s*)(第)(.*)(章)(.*)$|^(\\s*)(\\d)(\\s*)")
		result := reg.FindAllStringSubmatch(s,-1)
		a := result[0][0]
		return a
	}
	return ""
}

func getbookid(fp string) (bookid int,bookstorepath string){
	db := dbcon

	bn := strings.Split(filepath.Base(fp), ".")
	bookname := bn[2]

	err := db.QueryRow("SELECT id FROM bookinfo WHERE bookname = ?", bookname).Scan(&bookid)
	check(err)

	err = db.QueryRow("SELECT bookstorepath FROM bookinfo WHERE bookname = ?", bookname).Scan(&bookstorepath)
	check(err)

	return
}

func (c *chapterinfo)writetodb()  {
	db := dbcon

	stmt, err := db.Prepare(`INSERT chapter ( bookid, chaptername, chapter,storepath,chapterlines,chapterid) VALUES (?,?,?,?,?,?)`)
	check(err)

	res, err := stmt.Exec(c.Bookid,c.Chaptername,c.Chapter,c.Storepath,c.Chapterlines,c.Chapterid)
	check(err)

	id, err := res.LastInsertId() //必须是自增id的才可以正确返回。
	check(err)
	idstr := fmt.Sprintf("%v", id)

	fmt.Println(idstr)
}

func (c *chapterinfo)printcahpter()  {
	fmt.Println(c.Chaptername,c.Chapterlines)
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}


func GetAllFile(pathname string, fn_ch chan string) {
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		fmt.Println("read dir fail:", err)
	}
	for _, fi := range rd {
		if fi.IsDir() {
			fullDir := pathname + "/" + fi.Name()
			GetAllFile(fullDir, fn_ch)
			if err != nil {
				fmt.Println("read dir fail:", err)
			}
		} else {
			fullName := pathname + "/" + fi.Name()
			fn_ch <- fullName
		}
	}
}
