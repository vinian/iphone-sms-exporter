package main

import (
	"code.google.com/p/go-sqlite/go1/sqlite3"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
)

func main() {
	current_dir, err := user.Current()
	if err != nil {
		fmt.Println("main")
		log.Fatal(err)
	}
	pathinfo := []string{current_dir.HomeDir, "AppData", "Roaming", "Apple Computer", "MobileSync", "Backup"}
	backup_path := make_path(pathinfo)
	filepath.Walk(backup_path, file_func)
}

func make_path(pathinfo []string) string {
	//	str := strings.Join(pathinfo, string(os.PathSeparator))
	str := path.Join(pathinfo...)
	return str
}

func file_func(path string, info os.FileInfo, err error) error {
	if info.IsDir() || info.Size() == 0 {
		return nil
	}

	if is_sqlite_file(path) == 1 {
		fmt.Println(path)
		get_info_from_file(path)
	}

	return err
}

func is_sqlite_file(filename string) int {
	//	fmt.Println(filename)
	fileHandler, err := os.Open(filename)
	if err != nil {
		fmt.Println("is_sqlite_file open file error")
		log.Fatal(err)
	}
	defer fileHandler.Close()

	data := make([]byte, 20)
	if _, err := fileHandler.Read(data); err != nil {
		fmt.Println("is sqlite file read data err")
		log.Fatal(err)
	}

	flag := 0
	//	fmt.Println(string(data))
	matchstring, _ := regexp.Compile("SQLite format")
	if matchstring.Match(data) {
		flag = 1
	}

	return flag
}

func get_info_from_file(path string) {
	dbh, err := sqlite3.Open(path)
	if err != nil {
		fmt.Println("get_info_from_file")
		log.Fatal(err)
	}

	defer dbh.Close()

	if exists_message_table(dbh) == 1 {
		get_message_from_db(dbh)
	}
}

func exists_message_table(dbh *sqlite3.Conn) int {
	sql := "SELECT name FROM sqlite_master WHERE type='table' AND name='message';"

	flag := 0
	exists, err := dbh.Query(sql)

	if err == nil {
		if exists.Valid() {
			fmt.Println(exists.String())
			flag = 1
		}
	}

	return flag
}

func get_message_from_db(dbh *sqlite3.Conn) {
	sql := "select chat.guid, message.text from chat, message where chat.account_id=message.account_guid order by message.date;"

	out_file, err := os.OpenFile("./sms.export", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println("open sms out file error")
		log.Fatal(err)
	}
	defer out_file.Close()

	row := make(sqlite3.RowMap)
	for data, err := dbh.Query(sql); err == nil; err = data.Next() {
		data.Scan(row)
		if str, ok := row["text"].(string); ok {
			if _, err := out_file.WriteString(str); err != nil {
				fmt.Println("write sms file error")
				log.Fatal(err)
			}
			out_file.WriteString("\n")
		} else {
			fmt.Println("interface to string error")
			fmt.Println(row)
		}
	}
}

// for my friend Alice

