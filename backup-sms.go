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
	"runtime"
	"strings"
	"time"
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

	fmt.Println("the export sms file lies on the Desktop and name is sms_backup.txt")
	time.Sleep(10000 * time.Millisecond)
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
		//		fmt.Println(path)
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
			//			fmt.Println("found message db file")
			flag = 1
		}
	}

	return flag
}

func get_message_from_db(dbh *sqlite3.Conn) {
	sql := `select chat.guid, message.text, message.date + 978307200 as m_date
        from chat, message where chat.account_id=message.account_guid order by
        chat.guid, message.date;`

	fmt.Println("Begin exporting ...")
	//	fmt.Println(sql)
	sms_backup_file := get_backup_file()
	out_file, err := os.OpenFile(sms_backup_file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println("open sms out file error")
		log.Fatal(err)
	}
	defer out_file.Close()

	row := make(sqlite3.RowMap)
	for data, err := dbh.Query(sql); err == nil; err = data.Next() {
		data.Scan(row)
		if str, ok := row["text"].(string); ok {
			date_now := row["m_date"].(int64)
			sms_time := time.Unix(date_now, 0)
			phone_number := get_phone_number(row["guid"])
			sms := phone_number + " " + sms_time.String() + " " + str
			if _, err := out_file.WriteString(sms); err != nil {
				fmt.Println("write sms file error")
				log.Fatal(err)
			}
			if strings.EqualFold(runtime.GOOS, "windows") {
				out_file.WriteString("\r\n")
			} else {
				out_file.WriteString("\n")
			}
		} else {
			fmt.Println("interface to string error")
			fmt.Println(row)
		}
	}

	fmt.Println("Export OK")
}

func get_phone_number(sms_number interface{}) string {
	if str, ok := sms_number.(string); ok {
		number_format, _ := regexp.Compile(`\d+`)
		phone := number_format.FindString(str)
		return phone
	}

	return "unknown"
}

func get_backup_file() string {
	current_dir, err := user.Current()
	if err != nil {
		fmt.Println("get backup file error")
		log.Fatal(err)
	}
	pathinfo := []string{current_dir.HomeDir, "Desktop", "sms_backup.txt"}
	backup_file := make_path(pathinfo)

	return backup_file
}

// for my friend Alice
