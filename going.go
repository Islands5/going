package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/subcommands"
	"gopkg.in/yaml.v2"
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&initCmd{}, "")
	subcommands.Register(&upCmd{}, "")
	subcommands.Register(&resetCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}

///// init /////
type initCmd struct {
}

func (*initCmd) Name() string     { return "init" }
func (*initCmd) Synopsis() string { return "init" }
func (*initCmd) Usage() string {
	return `init:
  it makes files to use other commands.
`
}

func (p *initCmd) SetFlags(f *flag.FlagSet) {
}

func (p *initCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	path := os.Getenv("GOPATH") + "/src/github.com/islands5/going/assets"
	err := exec.Command("cp", "-r", path, "./going-assets").Run()
	if err != nil {
		fmt.Println(err)
	}
	return subcommands.ExitSuccess
}

///// up /////
type upCmd struct {
}

func (*upCmd) Name() string     { return "up" }
func (*upCmd) Synopsis() string { return "execute db migration" }
func (*upCmd) Usage() string {
	return `up:
  it executes db migration and records finished *.sql.
`
}

func (p *upCmd) SetFlags(f *flag.FlagSet) {
}

func (p *upCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	var assetsPath = "going-assets"
	connInfo, err := loadYml(assetsPath + "/going.yml")

	db, err := connectMysql(connInfo["db_name"], connInfo["user"], connInfo["password"])
	defer db.Close()

	files, err := ioutil.ReadDir(assetsPath + "/sql")
	if err != nil {
		fmt.Println(err)
	}

	var fileName string
	for _, f := range files {
		fileName = f.Name()
		reg := regexp.MustCompile(`(V[0-9].*?)__.*.sql`)
		match := reg.FindSubmatch([]byte(fileName))
		if match == nil {
			fmt.Println("Migration file format isn't correct. so please fix V{version}__ prefix.")
		}

		skip, err := isApplied(assetsPath, match[1])
		if skip {
			fmt.Println("already applied: " + fileName)
			continue
		}
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(fmt.Sprintf("applying: %s...", fileName))

		execSQL(db, assetsPath+"/sql/"+fileName)

		recordGoing(assetsPath, match[1])
	}

	return subcommands.ExitSuccess
}

///// reset /////
type resetCmd struct {
}

func (*resetCmd) Name() string     { return "reset" }
func (*resetCmd) Synopsis() string { return "db clear" }
func (*resetCmd) Usage() string {
	return `reset:
  it clear db and delete record files.
`
}

func (p *resetCmd) SetFlags(f *flag.FlagSet) {
}

func (p *resetCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	var assetsPath = "going-assets"
	connInfo, err := loadYml(assetsPath + "/going.yml")

	db, err := connectMysql("", "root", "password")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	db.Query(fmt.Sprintf("DROP DATABASE %s", connInfo["db_name"]))
	db.Query(fmt.Sprintf("CREATE DATABASE %s", connInfo["db_name"]))

	err = os.Remove(assetsPath + "/.going")
	if err != nil {
		fmt.Println(err)
	}

	err = exec.Command("touch", assetsPath+"/.going").Run()
	if err != nil {
		fmt.Println(err)
	}

	return subcommands.ExitSuccess
}

///// util /////
func connectMysql(database, user, password interface{}) (*sql.DB, error) {
	conn := fmt.Sprintf("%s:%s@/%s", user, password, database)

	db, err := sql.Open("mysql", conn)
	if err != nil {
		fmt.Println(err)
	}

	err = db.Ping()
	if err != nil {
		fmt.Println(err)
	}

	return db, err
}

func loadYml(filename string) (map[interface{}]interface{}, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}

	connInfo := make(map[interface{}]interface{})
	err = yaml.Unmarshal(buf, &connInfo)
	if err != nil {
		fmt.Println(err)
	}
	return connInfo, err
}

func execSQL(db *sql.DB, filename string) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}
	_, err = db.Query(string(buf))
	if err != nil {
		fmt.Println(err)
		return err
	}

	return err
}

func recordGoing(path string, version []byte) {
	f := string(version) + "__" + string(time.Now().Format("20060102030405")) + "\n"
	fp, err := os.OpenFile(path+"/.going", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		fmt.Println(err)
	}
	defer fp.Close()
	writer := bufio.NewWriter(fp)
	_, err = writer.WriteString(f)
	if err != nil {
		fmt.Println(err)
	}
	writer.Flush()
}

func isApplied(path string, version []byte) (bool, error) {
	buf, err := ioutil.ReadFile(path + "/.going")
	if err != nil {
		fmt.Println(err)
	}
	reg := regexp.MustCompile(string(version))

	return reg.Match(buf), err
}
