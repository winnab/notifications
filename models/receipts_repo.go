package models

import (
    "database/sql"
    "fmt"
    "strings"
    "time"
)

type ReceiptsRepo struct{}

type ReceiptsRepoInterface interface {
    CreateReceipts(ConnectionInterface, []string, string, string) error
}

func NewReceiptsRepo() ReceiptsRepo {
    return ReceiptsRepo{}
}

func (repo ReceiptsRepo) Create(conn ConnectionInterface, receipt Receipt) (Receipt, error) {
    receipt.CreatedAt = time.Now().Truncate(1 * time.Second).UTC()
    receipt.Count = 1
    err := conn.Insert(&receipt)
    if err != nil {
        if strings.Contains(err.Error(), "Duplicate entry") {
            err = ErrDuplicateRecord{}
        }
        return Receipt{}, err
    }

    return receipt, nil
}

func (repo ReceiptsRepo) Find(conn ConnectionInterface, userGUID, clientID, kindID string) (Receipt, error) {
    receipt := Receipt{}
    err := conn.SelectOne(&receipt, "SELECT * FROM  `receipts` WHERE `user_guid` = ? AND `client_id` = ? AND `kind_id` = ?", userGUID, clientID, kindID)
    if err != nil {
        if err == sql.ErrNoRows {
            err = ErrRecordNotFound{}
        }
        return Receipt{}, err
    }
    return receipt, nil
}

func (repo ReceiptsRepo) Update(conn ConnectionInterface, receipt Receipt) (Receipt, error) {
    _, err := conn.Update(&receipt)
    if err != nil {
        return receipt, err
    }

    return repo.Find(conn, receipt.UserGUID, receipt.ClientID, receipt.KindID)
}

func (repo ReceiptsRepo) CreateReceipts(conn ConnectionInterface, userGUIDs []string, clientID, kindID string) error {

    query := "INSERT INTO `receipts` (`user_guid`, `client_id`, `kind_id`, `count`, `created_at`) VALUES %s ON DUPLICATE KEY UPDATE `count`=`count`+1"
    values := ""
    var execArguments []interface{}

    for index, guid := range userGUIDs {
        receipt := Receipt{
            UserGUID: guid,
            ClientID: clientID,
            KindID:   kindID,
        }
        execArguments = append(execArguments, repo.buildExecArguments(receipt)...)
        values += "(?, ?, ?, ?, ?)"
        if index != len(userGUIDs)-1 {
            values += ","
        }
    }
    query = fmt.Sprintf(query, values)

    _, err := conn.Exec(query, execArguments...)
    if err != nil {
        return err
    }
    return nil
}

func (repo ReceiptsRepo) buildExecArguments(receipt Receipt) []interface{} {
    return []interface{}{receipt.UserGUID, receipt.ClientID, receipt.KindID, 1, time.Now().Truncate(1 * time.Second).UTC()}
}