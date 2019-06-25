package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
)

type Bookinfo struct {
	Bookid	int64	`db:"id"`
	Bookname string	`db:"bookname"`
	Boookauthor string `db:"boookauthor"`
	Bookcahtpernum	int `db:"bookcahtpernum"`
	Bookcomment	string `db:"bookcomment"`
	Bookstorepath string `db:"bookstorepath"`
}


func main() {
	b := Bookinfo{}
	b.getinfo()
	b.insert()
}

func (b *Bookinfo)getinfo()  {
	b.Bookname = "史上最强二道贩子"
	b.Boookauthor = "无我"
	b.Bookcahtpernum = 200
	b.Bookcomment = ""
	b.Bookstorepath = "/data/zip/unrarfull/史上最强二道贩子.txt"
}

func (b *Bookinfo)insert()  {
	db, err := sql.Open("mysql", "root:Yang@1008@tcp(192.168.1.28:3306)/book")
	check(err)

	stmt, err := db.Prepare(`INSERT bookinfo ( bookname, boookauthor, bookcahtpernum,Bookcomment,bookstorepath) VALUES (?,?,?,?,?)`)
	check(err)

	res, err := stmt.Exec(b.Bookname,b.Boookauthor,b.Bookcahtpernum,b.Bookcomment,b.Bookstorepath)
	check(err)

	id, err := res.LastInsertId()  //必须是自增id的才可以正确返回。
	check(err)

	fmt.Println(id)
	stmt.Close()
}

func check(err error) {
	if err != nil{
		fmt.Println(err)
		panic(err)
	}
}
