/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// FabricAsset implements a simple chaincode to manage an asset
type FabricAsset struct {
}

var string_HashValue string
var lockTime string
var lockPeriod string

// var lockMoney string
// var balance string

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *FabricAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {

	return shim.Success(nil)
}

func (t *FabricAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error
	if fn == "get" {
		result, err = get(stub, args)
	} else if fn == "add" {
		result, err = add(stub, args)
	} else if fn == "del" {
		result, err = del(stub, args)
	} else if fn == "lock" {
		result, err = lock(stub, args)
	} else if fn == "unlock" {
		result, err = unlock(stub, args)
	} else if fn == "rollback" {
		result, err = rollback(stub, args)
	} else if fn == "show" {
		result, err = show(stub, args)
	} else if fn == "InitAll" {
		result, err = InitAll(stub)
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(result))
}

func InitAll(stub shim.ChaincodeStubInterface) (string, error) {
	//插入jack
	m1 := map[string]string{
		"username": "jack",
		"balance":  "600",
	}

	josnValue, err := json.Marshal(m1)
	if err != nil {
		return "", err
	}

	err = stub.PutState("jack", josnValue)
	if err != nil {
		shim.Error("Init jack Error!")
	}

	m2 := map[string]string{
		"username": "jack1",
		"balance":  "600",
	}
	//插入jack1
	josnValue2, err := json.Marshal(m2)
	if err != nil {
		return "", err
	}

	err = stub.PutState("jack1", josnValue2)
	if err != nil {
		shim.Error("Init jack1 Error!")
	}

	//锁定状态
	statusLock := map[string]string{
		"status": "Null",
	}

	josnValue3, err := json.Marshal(statusLock)
	if err != nil {
		return "", err
	}

	err = stub.PutState("lock", josnValue3)
	if err != nil {
		shim.Error("Failed to Init lock state")
	}
	//锁定地址
	LockAdd := map[string]string{
		"status": "Null",
	}
	josnValue4, err := json.Marshal(LockAdd)
	if err != nil {
		return "", err
	}

	err = stub.PutState("lockAddress", josnValue4)
	if err != nil {
		shim.Error("Failed to Init lockAddress")
	}
	return "init compeleted !", nil
}

func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	return string(value), nil
}

func add(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a username and balance")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to add user add: %s", args[0])
	}

	m := map[string]string{
		"username": args[0],
		"balance":  args[1],
	}

	josnValue, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(josnValue), nil
}

func del(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a username")
	}

	err := stub.DelState(args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to del user add: %s", args[0])
	}
	m := map[string]string{
		"username": args[0],
		"status":   "deleted !",
	}

	josnValue, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(josnValue), nil
}

func lock(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 6 {
		return "Incorrect arguments.", fmt.Errorf("Incorrect arguments:%#v.", args)
	}
	//对哈希原像处理
	sha := sha256.New()
	hashValue := sha.Sum([]byte(args[2]))
	string_HashValue = hex.EncodeToString(hashValue)
	//获取锁定地址
	addr := "lockAddress"

	//锁定状态设置 转出账户,转入账户,哈希值,锁定资金,锁定时长，当前时间,锁定地址(这个参数不是外部传递进来的)
	//v := []byte(args[0] + "," + args[1] + "," + string_HashValue + "," + args[3] + "," + args[4] + "," + args[5] + string(addr))

	//序列化
	m_row := map[string]string{
		"out":         args[0],
		"in":          args[1],
		"Hash":        string_HashValue,
		"lockmoney":   args[3],
		"during":      args[4],
		"rightnow":    args[5],
		"lockAddress": string(addr),
	}

	josnV, err := json.Marshal(m_row)
	if err != nil {
		return "", err
	}

	err = stub.PutState("lock", josnV)
	if err != nil {
		return "", fmt.Errorf("Failed to add lock: %s", args[0])
	}

	lockPeriod = args[4]
	lockTime = args[5]
	//获取账户状态
	str_balance, err := stub.GetState(args[0])
	//balance = string(str_balance)
	if err != nil {
		return "", fmt.Errorf("Failed to get user balance: %s", args[0])
	}

	//比较金额大小
	int_balance, _ := strconv.Atoi(string(str_balance))
	//	lockMoney = args[3]
	int_lockMoney, _ := strconv.Atoi(args[3])
	if int_balance < int_lockMoney {
		return "balance insufficient!", fmt.Errorf("balance insufficient!")

	}

	//更新账户状态
	UpdateAccount := string(str_balance) + ",locked :" + args[3]
	err = stub.PutState(args[0], []byte(UpdateAccount))
	if err != nil {
		return "", fmt.Errorf("Failed to UpdateAccount: %s", args[0])
	}
	//更新锁定地址金额
	UpdatelockAddress := args[3]
	err = stub.PutState("lockAddress", []byte(UpdatelockAddress))
	if err != nil {
		return "Failed to update lockAddress", fmt.Errorf("balance insufficient!")
	}

	fmt.Println("money is locked:" + UpdateAccount)
	m := map[string]string{
		"addr": "lockAddress",
	}
	josnValue, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	//return args[0] + " Unlocked !", nil
	return string(josnValue), nil

}
func unlock(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 4 {
		return "", fmt.Errorf("Incorrect arguments:%#v.", args)
	}

	lockState, err := stub.GetState("lock")
	if err != nil {
		return "", fmt.Errorf("Failed to get lockstate in unlock.")
	}
	if string(lockState) == "Null" {
		return "", fmt.Errorf("There is no lock.")
	}
	//计算哈希值，并进行比较
	sha := sha256.New()
	hashValue := sha.Sum([]byte(args[1]))
	string_NewHashValue := hex.EncodeToString(hashValue)
	if string_NewHashValue != string_HashValue {
		return "Wrong serect!", fmt.Errorf("Wrong serect!")
	}
	//判断有没有过期
	int_lockTimeStart, _ := strconv.Atoi(lockTime)
	int_lockTimeEnd, _ := strconv.Atoi(args[3])
	int_lockPeriod, _ := strconv.Atoi(lockPeriod)

	if int_lockTimeEnd-int_lockTimeStart > int_lockPeriod {
		Args := make([]string, 0)
		Args = append(Args, args[0], args[1])
		//Unlock时超时回滚
		info, err := rollback(stub, Args)
		return "Lock is timeout ,start rollback :" + info, err
	}
	//更新unlock后的账户状态
	row_accountBalance, err := stub.GetState(args[0])
	// 比如"700,locked :100" 变为 "700"
	re := regexp.MustCompile(`^[^,]+`)
	s := string(row_accountBalance)
	accountBalance := re.FindString(s)

	if err != nil {
		return "Failed to get current balance in unlock", err
	}
	int_accountBalance, _ := strconv.Atoi(string(accountBalance))

	//匹配锁定的钱
	re_lockMoney := regexp.MustCompile(`[[:digit:]]+$`)

	LockMoney := re_lockMoney.FindString(string(row_accountBalance))

	int_lockMoney, _ := strconv.Atoi(LockMoney)

	int_accountBalanceUnlocked := int_accountBalance - int_lockMoney

	str_accountBalanceUnlocked := strconv.Itoa(int_accountBalanceUnlocked)

	err = stub.PutState(args[0], []byte(str_accountBalanceUnlocked))
	if err != nil {
		return "Failed to accountBalanceUnlocked", fmt.Errorf("Failed to accountBalanceUnlocked")
	}
	//更新锁定地址金额
	UpdatelockAddress := "Null"
	err = stub.PutState("lockAddress", []byte(UpdatelockAddress))
	if err != nil {
		return "Failed to update lockAddress", fmt.Errorf("Failed to update lockAddress")
	}
	//更新锁定状态
	err = stub.PutState("lock", []byte("Null"))
	if err != nil {
		return "", fmt.Errorf("Failed to back to lock state in unlock: %s", args[0])
	}
	//获取交易txid
	prove := stub.GetTxID()

	m := map[string]string{
		"hash":   prove,
		"time":   args[3],
		"action": "1",
		"status": "1",
	}
	josnValue, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	//return args[0] + " Unlocked !", nil
	return string(josnValue), nil
}

//没有使用到h原像
func rollback(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	//回滚更新账户
	// 比如"700,locked :100" 变为 "700"
	row_accountBalance, err := stub.GetState(args[0])
	re := regexp.MustCompile(`^[^,]+`)
	s := string(row_accountBalance)
	accountBalance := re.FindString(s)

	err = stub.PutState(args[0], []byte(accountBalance))
	if err != nil {
		return "Failed to rollback balance", fmt.Errorf("Failed to rollback balance , err:%#v.", err)
	}
	//更新锁定地址金额
	UpdatelockAddress := "Null"
	err = stub.PutState("lockAddress", []byte(UpdatelockAddress))
	if err != nil {
		return "Failed to update lockAddress in rollback", fmt.Errorf("Failed to update lockAddress in rollback")
	}
	//更新锁定状态
	err = stub.PutState("lock", []byte("Null"))
	if err != nil {
		return "", fmt.Errorf("Failed to rollback lock state: %s", args[0])
	}
	m := map[string]string{
		"data": "success !",
	}
	josnValue, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	//return args[0] + " rollback compeleted !", nil
	return string(josnValue), nil
}

func show(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	//获取账户锁定状态
	lockAccount, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("There is no lock: %s", err)
	}

	josnValue, err := json.Marshal(lockAccount)

	if err != nil {
		return "", err
	}

	return string(josnValue), nil
}

// main function starts up the chaincode in the container during instantiate
func main() {
	cur := time.Now().Unix()
	fmt.Println(cur, int(cur))
	if err := shim.Start(new(FabricAsset)); err != nil {
		fmt.Printf("Error starting FabricAsset chaincode: %s", err)
	}
}
