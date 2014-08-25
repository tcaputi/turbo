package turbo

import (
	"errors"
)

type DataTree struct {
	transactionMap	 map[string][]Transaction
	database	Database
}

type Transaction struct {
	connid	uint64
	txid	int64
	changed	bool
}

func NewDataTree(connString string, dbName, string, dbType, string) (*DataTree, error){
	db := NewDatabase(connString, dbName, dbType)
	var tm map[string][]Transaction
	tm = make(map[string][]Transaction)
	return DataTree{db, tm}
}

func (dt *DataTree) createTransaction(path string, connid uint64) int64{
	if dt.transactionMap[path] = nil {
		dt.transactionMap[path] = []Transaction
	}
	txid := time.Now().UnixNano()
	dt.transactionMap[path] = append(dt.transactionMap[path], Transaction{connid, txid, false})
} 

func (dt *DataTree) markTransactionsInvalid(path string){
	//TODO: this same process must be run for all parents
	if dt.transactionMap[path] = nil {
		return
	}
	for _, tx := range dt.transactionMap[path] {
		tx.changed = true
	}
}

func (dt *DataTree) attemptTransactionComplete(path string, connid uint64, txid int64) bool{
	for i, tx := range dt.transactionMap[path] {
		if tx.txid == txid && tx.connid == connid {
			if !tx.changed {
				tx.changed = 0
				return false
			} else {
				dt.transactionMap[path] = dt.transactionMap[path][:i+copy(dt.transactionMap[path][i:], dt.transactionMap[path][i+1:])]
				return true
			}
		}
	}
}

func (dt *DataTree) get(path string) (map[string]interface{}, error){
	return dt.database.get(path)
}

func (dt *DataTree) set(values map[string]interface{}) error{
	//update map changed -> true
	//TODO: is there a way i can do this only once without looping over the whole map?
	//delete direct parent values
	//TODO: write and expose this method in db.go
	return dt.database.set(values)
}

func (dt *DataTree) transget(path string, connid uint64) (map[string]interface{}, error){
	dt.createTransaction(path, connid)
	return dt.database.get(path)
}

func (dt *DataTree) transset(values map[string]interface{}) (map[string]interface{}, error){
	//if map entry changed = false //TODO make attemptTransactionComplete work here
		//remove entry from map
		return nil, dt.database.set(values)
	//else
		values, getErr := dt.database.get(path)
		if getErr != nil{
			return nil, getErr
		}
		return values, errors.New('conflict')
}