package adapter

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"strings"

	bolt "github.com/boltdb/bolt"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

// CasbinRule represents a Casbin rule line.
type CasbinRule []string

func (rule CasbinRule) getKey() string {
	return strings.Join(rule, ",")
}

type adapter struct {
	db            *bolt.DB
	namespace     []byte
	builtinPolicy string
}

// NewAdapter creates a new adapter.
func NewAdapter(store *BoltStore, bucket string, builtinPolicy string) (*adapter, error) {
	if bucket == "" {
		return nil, errors.New("must provide a namespace")
	}

	adapter := &adapter{
		db:            store.conn,
		namespace:     []byte(bucket),
		builtinPolicy: builtinPolicy,
	}

	if err := adapter.init(); err != nil {
		return nil, err
	}

	return adapter, nil
}

func (a *adapter) init() error {
	return a.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(a.namespace)
		return err
	})
}

// LoadPolicy performs a scan on the namespace and individually loads every line into the Casbin model.
// Not particularity efficient but should only be required on when you application starts up as this adapter can
// leverage auto-save functionality.
func (a *adapter) LoadPolicy(model model.Model) error {
	if a.builtinPolicy != "" {
		for _, line := range strings.Split(a.builtinPolicy, "\n") {
			if err := loadCsvPolicyLine(strings.TrimSpace(line), model); err != nil {
				return err
			}
		}
	}

	return a.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(a.namespace)

		return bucket.ForEach(func(k, v []byte) error {
			var line CasbinRule
			if err := json.Unmarshal(v, &line); err != nil {
				return err
			}
			loadPolicy(line, model)
			return nil
		})
	})
}

// SavePolicy is not supported for this adapter. Auto-save should be used.
func (a *adapter) SavePolicy(model model.Model) error {
	//TODO
	return errors.New("not supported: must use auto-save with this adapter")
}

// AddPolicy inserts or updates a rule.
func (a *adapter) AddPolicy(sec string, ptype string, rule []string) error {
	return a.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(a.namespace)

		line := convertRule(ptype, rule)

		bts, err := json.Marshal(line)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(line.getKey()), bts)
	})
}

// AddPolicies inserts or updates multiple rules by iterating over each one and inserting it into the namespace.
func (a *adapter) AddPolicies(sec string, ptype string, rules [][]string) error {
	return a.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(a.namespace)

		for _, r := range rules {

			line := convertRule(ptype, r)

			bts, err := json.Marshal(line)
			if err != nil {
				return err
			}

			if err := bucket.Put([]byte(line.getKey()), bts); err != nil {
				return err
			}
		}

		return nil
	})
}

// RemoveFilteredPolicy has an implementation that is slightly limited in that we can only find and remove elements
// using a policy line prefix.
//
// For example, if you have the following policy:
//     p, subject-a, action-a, get
//     p, subject-a, action-a, write
//     p, subject-b, action-a, get
//     p, subject-b, action-a, write
//
// The following would remove all subject-a rules:
//     enforcer.RemoveFilteredPolicy(0, "subject-a")
// The following would remove all subject-a rules that contain action-a:
//     enforcer.RemoveFilteredPolicy(0, "subject-a", "action-a")
//
// The following does not work and will return an error:
//     enforcer.RemoveFilteredPolicy(1, "action-a")
//
// This is because we use leverage Bolts seek and prefix to find an item by prefix.
// Once these keys are found we can iterate over and remove them.
// Each policy rule is stored as a row in Bolt: p::subject-a::action-a::get
func (a *adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	panic("not implement")
}

func (rule CasbinRule) filter() string {
	filter := strings.Join(rule, "::")
	return filter
}

// RemovePolicy removes a policy line that matches key.
func (a *adapter) RemovePolicy(sec string, ptype string, line []string) error {
	return a.db.Update(func(tx *bolt.Tx) error {
		rule := convertRule(ptype, line)
		bucket := tx.Bucket(a.namespace)
		return bucket.Delete([]byte(rule.getKey()))
	})
}

// UpdateFilteredPolicies deletes old rules and adds new rules.
func (a *adapter) UpdateFilteredPolicies(sec string, ptype string, newPolicies [][]string, fieldIndex int, fieldValues ...string) ([][]string, error) {
	panic("implement")
}

// UpdatePolicy updates a policy rule from storage.
// This is part of the Auto-Save feature.
func (a *adapter) UpdatePolicy(sec string, ptype string, oldRule, newPolicy []string) error {
	return a.db.Update(func(tx *bolt.Tx) error {
		rule := convertRule(ptype, oldRule)
		bucket := tx.Bucket(a.namespace)
		if bucket.Get([]byte(rule.getKey())) != nil {
			if err := bucket.Delete([]byte(rule.getKey())); err != nil {
				return err
			}
			line := convertRule(ptype, newPolicy)
			bts, err := json.Marshal(line)
			if err != nil {
				return err
			}
			if err := bucket.Put([]byte(line.getKey()), bts); err != nil {
				return err
			}
		}
		return nil
	})
}

var (
	InvalidRulesLen = errors.New("invalid rule len")
)

// UpdatePolicies updates some policy rules to storage, like db, redis.
func (a *adapter) UpdatePolicies(sec string, ptype string, oldRules, newRules [][]string) error {
	return a.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(a.namespace)
		if len(oldRules) != len(newRules) {
			return InvalidRulesLen
		}
		for i := 0; i < len(oldRules); i++ {
			or := convertRule(ptype, oldRules[i])
			if bucket.Get([]byte(or.getKey())) != nil {
				if err := bucket.Delete([]byte(or.getKey())); err != nil {
					return err
				}
				line := convertRule(ptype, newRules[i])
				bts, err := json.Marshal(line)
				if err != nil {
					return err
				}
				if err := bucket.Put([]byte(line.getKey()), bts); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// RemovePolicies removes multiple policies.
func (a *adapter) RemovePolicies(sec string, ptype string, rules [][]string) error {
	return a.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(a.namespace)
		for _, r := range rules {
			rule := convertRule(ptype, r)
			if err := bucket.Delete([]byte(rule.getKey())); err != nil {
				return err
			}
		}
		return nil
	})
}

func loadPolicy(rule CasbinRule, model model.Model) {
	//lineText := rule.PType
	lineText := strings.Join(rule, ",")
	persist.LoadPolicyLine(lineText, model)
}

func loadCsvPolicyLine(line string, model model.Model) error {
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}

	reader := csv.NewReader(strings.NewReader(line))
	reader.TrimLeadingSpace = true
	tokens, err := reader.Read()
	if err != nil {
		return err
	}

	key := tokens[0]
	sec := key[:1]
	model[sec][key].Policy = append(model[sec][key].Policy, tokens[1:])
	return nil
}

func convertRule(ptype string, line []string) CasbinRule {
	keySlice := []string{ptype}
	keySlice = append(keySlice, line...)
	return keySlice
}
