package graphb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/pkg/errors"
	"fmt"
)

func TestTheWholePackage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		q := Query{
			OperationType: "query",
			OperationName: "",
			Fields: []*Field{
				{
					Name:  "courses",
					Alias: "Alias",
					Arguments: []Argument{
						ArgumentInt("uid", 123),
						ArgumentStringSlice("blocked_nds", "nd013", "nd014"),
					},
					Fields: Fields("key", "id"),
				},
			},
		}
		strCh, err := q.StringChan()
		str := StringFromChan(strCh)
		assert.Nil(t, err)
		assert.Equal(t, `query{Alias:courses(uid:123,blocked_nds:["nd013","nd014"]){key,id,},}`, str)

		str, err = q.JsonBody()
		assert.Nil(t, err)
		assert.Equal(t, `{"query":"query{Alias:courses(uid:123,blocked_nds:[\"nd013\",\"nd014\"]){key,id,},}"}`, str)

		strCh, err = q.Fields[0].StringChan()
		str = StringFromChan(strCh)
		assert.Nil(t, err)
		assert.Equal(t, `Alias:courses(uid:123,blocked_nds:["nd013","nd014"]){key,id,}`, str)
	})

	t.Run("Invalid names", func(t *testing.T) {
		q := Query{
			OperationType: "query",
			OperationName: "test_graphb",
			Fields: []*Field{
				{
					Name:      "courses",
					Alias:     "Lets_Have_An_Alias看",
					Arguments: nil,
					Fields:    Fields("id", "key"),
				},
			},
		}
		strCh, err := q.StringChan()
		assert.Equal(t, "'Lets_Have_An_Alias看' is an invalid name identifier in GraphQL. A valid name matches /[_A-Za-z][_0-9A-Za-z]*/, see: http://facebook.github.io/graphql/October2016/#sec-Names", err.Error())
		value, ok := <-strCh
		assert.Equal(t, "", value)
		assert.Equal(t, false, ok)
	})

	t.Run("check cycles", func(t *testing.T) {
		f := &Field{
			Name:  "courses",
			Alias: "Alias",
			Arguments: []Argument{
				ArgumentInt("uid", 123),
				ArgumentStringSlice("blocked_nds", "nd013", "nd014"),
			},
			Fields: Fields(""),
		}
		f2 := &Field{Fields: Fields("")}
		f2.Fields[0] = f
		f.Fields[0] = f2

		q := Query{
			OperationType: "query",
			OperationName: "",
			Fields:        []*Field{f2},
		}
		strCh, err := q.StringChan()
		assert.IsTypef(t, CyclicFieldErr{}, errors.Cause(err), "")
		value, ok := <-strCh
		assert.Equal(t, "", value)
		assert.Equal(t, false, ok)

		strCh, err = f.StringChan()
		assert.IsTypef(t, CyclicFieldErr{}, errors.Cause(err), "")
		value, ok = <-strCh
		assert.Equal(t, "", value)
		assert.Equal(t, false, ok)
	})

	t.Run("name validation", func(t *testing.T) {
		_, err := NewQuery(TypeQuery, OfName("我"))
		assert.Equal(t, "'我' is not a valid name.", err.Error())

		_, err = NewQuery(TypeQuery, OfName("_我"))
		assert.Equal(t, "'_我' is not a valid name.", err.Error())

		_, err = NewQuery(TypeMutation, OfName("x-x"))
		assert.Equal(t, "'x-x' is not a valid name.", err.Error())

		_, err = NewQuery(TypeMutation, OfName("x x"))
		assert.Equal(t, "'x x' is not a valid name.", err.Error())

		_, err = NewQuery(TypeSubscription, OfName("_1x1_1x1_"))
		assert.Nil(t, err)
	})

	t.Run("Nested fields", func(t *testing.T) {
		q, err := NewQuery(
			TypeQuery,
			OfName("another_test"),
			OfField(
				"users",
				OfFields("id", "username"),
				OfField(
					"threads",
					OfArguments(ArgumentString("title", "A Good Title")),
					OfFields("title", "created_at"),
				),
			),
		)
		assert.Nil(t, err)
		s, err := q.JsonBody()
		assert.Nil(t, err)
		assert.Equal(t, `{"query":"query another_test{users{id,username,threads(title:\"A Good Title\"){title,created_at,},},}"}`, s)
	})

	t.Run("Invalid operation type", func(t *testing.T) {
		q := Query{OperationType: "muTatio"}
		ch, err := q.StringChan()
		assert.Equal(t, "'muTatio' is an invalid operation type in GraphQL. A valid type is one of 'query', 'mutation', 'subscription'", err.Error())
		s, ok := <-ch
		assert.Equal(t, "", s)
		assert.False(t, ok)

		s, err = q.JsonBody()
		assert.Equal(t, "'muTatio' is an invalid operation type in GraphQL. A valid type is one of 'query', 'mutation', 'subscription'", err.Error())
		assert.Equal(t, "", s)
	})

	t.Run("Nil field error", func(t *testing.T) {
		q := Query{OperationType: "mutation", Fields: []*Field{nil}}
		ch, err := q.StringChan()
		assert.Equal(t, "nil Field is not allowed. Please initialize a correct Field with NewField(...) function or Field{...} literal", err.Error())
		s, ok := <-ch
		assert.Equal(t, "", s)
		assert.False(t, ok)
	})

	t.Run("Nil field error 2", func(t *testing.T) {
		f := Field{Fields: []*Field{nil}}
		err := f.checkCycle()
		assert.IsTypef(t, NilFieldErr{}, errors.Cause(err), "")
	})
}

func TestMethodChaining(t *testing.T) {
	t.Run("", func(t *testing.T) {
		q := MakeQuery(TypeQuery).SetOperationName("").SetFields(
			MakeField().AddArgumentString("uid", "123"),
		)
		s, _ := q.JsonBody()
		fmt.Println(s)
	})
}