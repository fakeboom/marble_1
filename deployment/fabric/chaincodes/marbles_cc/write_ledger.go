package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"reflect"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// ============================================================================================================================
// write() - genric write variable into ledger
//
// Shows Off PutState() - writting a key/value into the ledger
//
// Inputs - Array of strings
//    0   ,    1
//   key  ,  value
//  "abc" , "test"
// ============================================================================================================================
func write(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key, value string
	var err error
	fmt.Println("starting write")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2. key of the variable and value to set")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	key = args[0] //rename for funsies
	value = args[1]
	err = stub.PutState(key, []byte(value)) //write the variable into the ledger
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end write")
	return shim.Success(nil)
}

// 输入id删除对应的值
func delete_marble(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("starting delete_marble")

	if len(args) != 1{
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// input sanitation
	err := sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	id := args[0]
	err = stub.DelState(id) //remove the key from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	fmt.Println("- end delete_marble")
	return shim.Success(nil)
}

// ============================================================================================================================
// Init Marble - create a new marble, store into chaincode state
//
// Shows off building a key's JSON value manually
//
// Inputs - Array of strings
//      0      ,    1  ,  2  ,      3          ,       4         ,        5
//     id      ,  color, size,     owner id    ,  authing company, additional data
// "m999999999", "blue", "35", "o9999999999999", "united marbles", "hbas76asdjhg67"
// ============================================================================================================================
func init_marble(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting init_marble")

	if len(args) < 5 || len(args) > 6 {
		return shim.Error("Incorrect number of arguments. Expecting 5 or 6")
	}

	//input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	id := args[0]
	color := strings.ToLower(args[1])
	owner_id := args[3]
	authed_by_company := args[4]
	size, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	//check if new owner exists
	owner, err := get_owner(stub, owner_id)
	if err != nil {
		fmt.Println("Failed to find owner - " + owner_id)
		return shim.Error(err.Error())
	}

	//check authorizing company (see note in set_owner() about how this is quirky)
	if owner.Company != authed_by_company {
		return shim.Error("The company '" + authed_by_company + "' cannot authorize creation for '" + owner.Company + "'.")
	}

	//check if marble id already exists
	marble, err := get_marble(stub, id)
	if err == nil {
		fmt.Println("This marble already exists - " + id)
		fmt.Println(marble)
		return shim.Error("This marble already exists - " + id) //all stop a marble by this id exists
	}

	// check if there was additional data provided. Additional data is added only
	// to test the impact of content size on the blockchain
	var additionalData string = ""
	if len(args) == 6 {
		additionalData = args[5]
	}

	//build the marble json string manually
	str := `{
		"docType":"marble", 
		"id": "` + id + `", 
		"color": "` + color + `", 
		"size": ` + strconv.Itoa(size) + `, 
		"owner": {
			"id": "` + owner_id + `", 
			"username": "` + owner.Username + `", 
			"company": "` + owner.Company + `"
		},
		"additionalData": "` + additionalData + `"
	}`
	err = stub.PutState(id, []byte(str)) //store marble with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end init_marble")
	return shim.Success(nil)
}

func change(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting change"+args[0])


	//input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	id := args[0]
	var res interface{}
	var strtype string 
	switch id[0]{
		case 'c' : res = &City{}
			strtype = "city"
		case 'i' : res = &Institution{}
			strtype = "institution"
		case 'e' : res = &Expert{}
			strtype = "expert"
		case 'm' : res = &Marble{}
			strtype = "marble"
		case 'd' : res = &Demand{}
			strtype = "demand"
		case 's' : res = &Scheme{}
			strtype = "scheme"
		case 'P' : res = &Patent{}
			strtype = "patent"
		case 'p' : res = &Paper{}
			strtype = "paper"
		case 't' : res = &Transfer{}
			strtype =  "transfer"
		default  :res = &Marble{}
			strtype = "marble"
	}
	rVal := reflect.ValueOf(res).Elem()
	for i:= 0 ; i<rVal.NumField();i++{
		f := rVal.Field(i)
		if i == 0{
			v:= strtype
			f.Set(reflect.ValueOf(v))
		}else {
			v:= args[i-1]
			f.Set(reflect.ValueOf(v))
		}
	}
	str, _ := json.Marshal(res)
	
	err = stub.PutState(id, str) //store marble with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end change")
	return shim.Success(nil)
}

// ============================================================================================================================
// Init Owner - create a new owner aka end user, store into chaincode state
//
// Shows off building key's value from GoLang Structure
//
// Inputs - Array of Strings
//           0     ,     1   ,   2
//      owner id   , username, company
// "o9999999999999",     bob", "united marbles"
// ============================================================================================================================
func init_owner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting init_owner")

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	//input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var owner Owner
	owner.ObjectType = "marble_owner"
	owner.Id = args[0]
	owner.Username = strings.ToLower(args[1])
	owner.Company = args[2]
	owner.Enabled = true
	fmt.Println(owner)

	//check if user already exists
	_, err = get_owner(stub, owner.Id)
	if err == nil {
		fmt.Println("This owner already exists - " + owner.Id)
		return shim.Error("This owner already exists - " + owner.Id)
	}

	//store user
	ownerAsBytes, _ := json.Marshal(owner)      //convert to array of bytes
	err = stub.PutState(owner.Id, ownerAsBytes) //store owner by its Id
	if err != nil {
		fmt.Println("Could not store user")
		return shim.Error(err.Error())
	}

	fmt.Println("- end init_owner marble")
	return shim.Success(nil)
}

//args[0] 要修改的id， args[1] 新的ownerid
func set_owner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting set_owner")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var marble_id = args[0]
	var new_owner_id = args[1]
	fmt.Println(marble_id + "->" + new_owner_id)

	// get marble's current state
	marbleAsBytes, err := stub.GetState(marble_id)
	if err != nil {
		return shim.Error("Failed to get marble")
	}


	switch marble_id[0]{
		case 'm' : res := Marble{}
				json.Unmarshal(marbleAsBytes, &res) 
				res.Owner.Id = new_owner_id
				jsonAsBytes, _ := json.Marshal(res)       
				err = stub.PutState(args[0], jsonAsBytes) 
		case 'd' : res := Demand{}
				json.Unmarshal(marbleAsBytes, &res) 
				res.OwnerId = new_owner_id
				jsonAsBytes, _ := json.Marshal(res)       
				err = stub.PutState(args[0], jsonAsBytes) 
		case 's' : res := Scheme{}
			json.Unmarshal(marbleAsBytes, &res) 
			res.OwnerId = new_owner_id
			jsonAsBytes, _ := json.Marshal(res)       
			err = stub.PutState(args[0], jsonAsBytes) 
		case 'P' : res := Patent{}
			json.Unmarshal(marbleAsBytes, &res) 
			res.OwnerId = new_owner_id
			jsonAsBytes, _ := json.Marshal(res)       
			err = stub.PutState(args[0], jsonAsBytes) 
		case 'p' : res := Paper{}
			json.Unmarshal(marbleAsBytes, &res) 
			res.OwnerId = new_owner_id
			jsonAsBytes, _ := json.Marshal(res)       
			err = stub.PutState(args[0], jsonAsBytes) 
		default  :res := Marble{}
			json.Unmarshal(marbleAsBytes, &res) 
			res.Owner.Id = new_owner_id
			jsonAsBytes, _ := json.Marshal(res)       
			err = stub.PutState(args[0], jsonAsBytes) 
	}

	
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end set owner")
	return shim.Success(nil)
}

func able_alkind(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting able")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var marble_id = args[0]
	fmt.Println(marble_id )

	// get marble's current state
	marbleAsBytes, err := stub.GetState(marble_id)
	if err != nil {
		return shim.Error("Failed to get marble")
	}


	switch marble_id[0]{
		case 'd' : res := Demand{}
				json.Unmarshal(marbleAsBytes, &res) 
				res.Able = "true"
				jsonAsBytes, _ := json.Marshal(res)       
				err = stub.PutState(args[0], jsonAsBytes) 
		case 's' : res := Scheme{}
			json.Unmarshal(marbleAsBytes, &res) 
			res.Able = "true"
			jsonAsBytes, _ := json.Marshal(res)       
			err = stub.PutState(args[0], jsonAsBytes) 
		case 'P' : res := Patent{}
			json.Unmarshal(marbleAsBytes, &res) 
			res.Able = "true"
			jsonAsBytes, _ := json.Marshal(res)       
			err = stub.PutState(args[0], jsonAsBytes) 
		case 'p' : res := Paper{}
			json.Unmarshal(marbleAsBytes, &res) 
			res.Able = "true"
			jsonAsBytes, _ := json.Marshal(res)       
			err = stub.PutState(args[0], jsonAsBytes) 
		case  'c' : res := City{}
			json.Unmarshal(marbleAsBytes, &res) 
			res.Able = "true"
			jsonAsBytes, _ := json.Marshal(res)       
			err = stub.PutState(args[0], jsonAsBytes)
		case  'i' :  res := Institution{}
			json.Unmarshal(marbleAsBytes, &res) 
			res.Able = "true"
			jsonAsBytes, _ := json.Marshal(res)       
			err = stub.PutState(args[0], jsonAsBytes)   
		case   'e' : res := Expert{}
			json.Unmarshal(marbleAsBytes, &res) 
			res.Able = "true"
			jsonAsBytes, _ := json.Marshal(res)       
			err = stub.PutState(args[0], jsonAsBytes) 
		case   't' : res := Transfer{}
			json.Unmarshal(marbleAsBytes, &res) 
			res.Able = "true"
			jsonAsBytes, _ := json.Marshal(res)       
			err = stub.PutState(args[0], jsonAsBytes) 
	}

	
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end able")
	return shim.Success(nil)
}

// ============================================================================================================================
// Disable Marble Owner
//
// Shows off PutState()
//
// Inputs - Array of Strings
//       0     ,        1
//  owner id       , company that auth the transfer
// "o9999999999999", "united_mables"
// ============================================================================================================================
func disable_owner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting disable_owner")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var owner_id = args[0]
	var authed_by_company = args[1]

	// get the marble owner data
	owner, err := get_owner(stub, owner_id)
	if err != nil {
		return shim.Error("This owner does not exist - " + owner_id)
	}

	// check authorizing company
	if owner.Company != authed_by_company {
		return shim.Error("The company '" + authed_by_company + "' cannot change another companies marble owner")
	}

	// disable the owner
	owner.Enabled = false
	jsonAsBytes, _ := json.Marshal(owner)     //convert to array of bytes
	err = stub.PutState(args[0], jsonAsBytes) //rewrite the owner
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end disable_owner")
	return shim.Success(nil)
}

// ============================================================================================================================
// Clear all marbles
//
// ============================================================================================================================
func clear_marbles(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("starting clear_marbles")
	resultsIterator, err := stub.GetStateByRange("m ", "m~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	mIDs := []string{}
	for resultsIterator.HasNext() {
		aKeyValue, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		mIDs = append(mIDs, aKeyValue.Key)
	}
	resultsIterator.Close()

	var delCount int
	for _, id := range mIDs {
		if err := stub.DelState(id); err != nil {
			fmt.Println("failed to delete marble: %s: %s", id, err)
		} else {
			delCount += 1
		}
	}
	fmt.Println("- end clear_marbles - found: %d, deleted: %d", len(mIDs))
	return shim.Success([]byte(fmt.Sprintf(`{"found": %d, "deleted": %d}`, len(mIDs), delCount)))
}

// ============================================================================================================================
// delete_marble_noauth() - delete a marble without checking auth company
//
// Inputs - Array of strings
//      0
//     id
// "m999999999"
// ============================================================================================================================
func delete_marble_noauth(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("starting delete_marble_noauth")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// input sanitation
	err := sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	id := args[0]

	// get the marble
	if _, err = get_marble(stub, id); err != nil {
		fmt.Println("Failed to find marble by id " + id)
		return shim.Error(err.Error())
	}

	// remove the marble
	err = stub.DelState(id) //remove the key from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	fmt.Println("- end delete_marble_noauth")
	return shim.Success(nil)
}
