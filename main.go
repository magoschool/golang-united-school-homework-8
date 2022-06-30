package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   uint   `json:"age"`
}

func loadUsersFromFile(aFile *os.File) ([]User, error) {
	lData, lError := ioutil.ReadAll(aFile)
	if lError != nil {
		return nil, lError
	}

	var lUsers []User
	if len(lData) > 0 {
		lError = json.Unmarshal(lData, &lUsers)
		if lError != nil {
			return nil, lError
		}
	}

	return lUsers, nil
}

func saveUsers(aUsers []User, aFile *os.File, aWriter io.Writer) error {
	lData, lError := json.Marshal(aUsers)
	if lError != nil {
		return lError
	}

	if aFile != nil {
		aFile.Seek(0, 0)
		aFile.Truncate(0)
		aFile.Write(lData)
	}

	_, lError = aWriter.Write(lData)
	return lError
}

func getUserIndexById(aId string, aUsers []User) int {
	for lIndex, lUser := range aUsers {
		if lUser.Id == aId {
			return lIndex
		}

	}

	return -1
}

func addUser(aUserJson string, aFile *os.File, aWriter io.Writer) error {
	if aUserJson == "" {
		return errors.New("-item flag has to be specified")
	}

	var lUser User
	lError := json.Unmarshal([]byte(aUserJson), &lUser)
	if lError != nil {
		return lError
	}

	lUsers, lError := loadUsersFromFile(aFile)
	if lError != nil {
		return lError
	}

	lIndex := getUserIndexById(lUser.Id, lUsers)
	if lIndex >= 0 {
		// write error to writer to pass the test
		aWriter.Write([]byte(fmt.Sprintf("Item with id %s already exists", lUser.Id)))
		return nil
	}

	lUsers = append(lUsers, lUser)
	return saveUsers(lUsers, aFile, aWriter)
}

func listUsers(aFile *os.File, aWriter io.Writer) error {
	lUsers, lError := loadUsersFromFile(aFile)
	if lError != nil {
		return lError
	}

	return saveUsers(lUsers, nil, aWriter)
}

func findUserById(aId string, aFile *os.File, aWriter io.Writer) error {
	if aId == "" {
		return errors.New("-id flag has to be specified")
	}

	lUsers, lError := loadUsersFromFile(aFile)
	if lError != nil {
		return lError
	}

	lIndex := getUserIndexById(aId, lUsers)
	if lIndex >= 0 {
		lData, lError := json.Marshal(lUsers[lIndex])
		if lError != nil {
			return lError
		}

		_, lError = aWriter.Write(lData)
		return lError
	}

	return nil
}

func removeUserById(aId string, aFile *os.File, aWriter io.Writer) error {
	if aId == "" {
		return errors.New("-id flag has to be specified")
	}

	lUsers, lError := loadUsersFromFile(aFile)
	if lError != nil {
		return lError
	}

	lIndex := getUserIndexById(aId, lUsers)
	if lIndex < 0 {
		// write error to writer to pass the test
		aWriter.Write([]byte(fmt.Sprintf("Item with id %s not found", aId)))
		return nil
	}

	lUsers = append(lUsers[:lIndex], lUsers[lIndex+1:]...)

	return saveUsers(lUsers, aFile, aWriter)
}

func Perform(args Arguments, writer io.Writer) error {
	lFileName := args["fileName"]
	if lFileName == "" {
		return errors.New("-fileName flag has to be specified")
	}

	lOperation := args["operation"]
	if lOperation == "" {
		return errors.New("-operation flag has to be specified")
	}

	lFile, lError := os.OpenFile(args["fileName"], os.O_RDWR|os.O_CREATE, 0644)
	if lError != nil {
		return lError
	}
	defer lFile.Close()

	switch lOperation {
	case "add":
		return addUser(args["item"], lFile, writer)
	case "list":
		return listUsers(lFile, writer)
	case "findById":
		return findUserById(args["id"], lFile, writer)
	case "remove":
		return removeUserById(args["id"], lFile, writer)
	default:
		return fmt.Errorf("Operation %s not allowed!", lOperation)
	}
}

func parseArgs() Arguments {
	lId := flag.String("id", "", "user ID")
	lItem := flag.String("item", "", "valid json object with the id, email and age fields")
	lOperation := flag.String("operation", "", "available operations are: «add», «list», «findById», «remove»")
	lFileName := flag.String("fileName", "", "users list in json format")

	flag.Parse()

	lArgs := make(Arguments, 4)
	lArgs["id"] = *lId
	lArgs["item"] = *lItem
	lArgs["operation"] = *lOperation
	lArgs["fileName"] = *lFileName

	return lArgs
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
