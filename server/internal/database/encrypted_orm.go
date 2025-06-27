package database

import (
	"reflect"

	"github.com/beego/beego/v2/client/orm"
	"github.com/datazip/olake-frontend/server/internal/crypto"
)

// EncryptableModel interface for models that have encrypted fields
type EncryptableModel interface {
	GetEncryptableFields() []string
	SetEncryptedField(fieldName, encryptedValue string) error
	GetEncryptedField(fieldName string) string
}

// EncryptedORM wraps the standard ORM with automatic encryption/decryption
type EncryptedORM struct {
	orm.Ormer
}

// NewEncryptedORM creates a new instance of EncryptedORM
func NewEncryptedORM() *EncryptedORM {
	return &EncryptedORM{
		Ormer: orm.NewOrm(),
	}
}

// Insert intercepts insert operations to encrypt fields
func (e *EncryptedORM) Insert(md interface{}) (int64, error) {
	if err := e.encryptFields(md); err != nil {
		return 0, err
	}
	return e.Ormer.Insert(md)
}

// Update intercepts update operations to encrypt fields
func (e *EncryptedORM) Update(md interface{}, cols ...string) (int64, error) {
	if err := e.encryptFields(md); err != nil {
		return 0, err
	}
	return e.Ormer.Update(md, cols...)
}

// Read intercepts read operations to decrypt fields
func (e *EncryptedORM) Read(md interface{}, cols ...string) error {
	if err := e.Ormer.Read(md, cols...); err != nil {
		return err
	}
	return e.decryptFields(md)
}

// QueryTable returns a QuerySeter with automatic decryption
func (e *EncryptedORM) QueryTable(ptrStructOrTableName interface{}) orm.QuerySeter {
	return &EncryptedQuerySeter{
		QuerySeter: e.Ormer.QueryTable(ptrStructOrTableName),
		orm:        e,
	}
}

// encryptFields encrypts the encryptable fields of a model
func (e *EncryptedORM) encryptFields(md interface{}) error {
	if encryptable, ok := md.(EncryptableModel); ok {
		fields := encryptable.GetEncryptableFields()
		for _, fieldName := range fields {
			plainText := encryptable.GetEncryptedField(fieldName)
			if plainText != "" {
				encryptedValue, err := crypto.EncryptJSONString(plainText)
				if err != nil {
					return err
				}
				if err := encryptable.SetEncryptedField(fieldName, encryptedValue); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// decryptFields decrypts the encryptable fields of a model
func (e *EncryptedORM) decryptFields(md interface{}) error {
	if encryptable, ok := md.(EncryptableModel); ok {
		fields := encryptable.GetEncryptableFields()
		for _, fieldName := range fields {
			encryptedValue := encryptable.GetEncryptedField(fieldName)
			if encryptedValue != "" {
				decryptedValue, err := crypto.DecryptJSONString(encryptedValue)
				if err != nil {
					return err
				}
				if err := encryptable.SetEncryptedField(fieldName, decryptedValue); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// EncryptedQuerySeter wraps QuerySeter with automatic decryption
type EncryptedQuerySeter struct {
	orm.QuerySeter
	orm *EncryptedORM
}

// All executes query and decrypts results
func (e *EncryptedQuerySeter) All(container interface{}, cols ...string) (int64, error) {
	num, err := e.QuerySeter.All(container, cols...)
	if err != nil {
		return num, err
	}

	// Decrypt all items in the container
	containerValue := reflect.ValueOf(container)
	if containerValue.Kind() == reflect.Ptr {
		containerValue = containerValue.Elem()
	}

	if containerValue.Kind() == reflect.Slice {
		for i := 0; i < containerValue.Len(); i++ {
			item := containerValue.Index(i)
			if item.Kind() == reflect.Ptr {
				if err := e.orm.decryptFields(item.Interface()); err != nil {
					return num, err
				}
			}
		}
	}

	return num, nil
}

// One executes query and decrypts result
func (e *EncryptedQuerySeter) One(ptrStruct interface{}, cols ...string) error {
	if err := e.QuerySeter.One(ptrStruct, cols...); err != nil {
		return err
	}
	return e.orm.decryptFields(ptrStruct)
}
