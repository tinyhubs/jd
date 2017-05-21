package jd

import (
	"container/list"
	"encoding/json"
	"fmt"
	"io"
)

const (
	TYPE_NULL   = 0x0001
	TYPE_BOOL   = 0x0002
	TYPE_NUMBER = 0x0004
	TYPE_STRING = 0x0010
	TYPE_ARRAY  = 0x0100
	TYPE_MAP    = 0x0200
)

type Object interface {
	Type() uint32

	ToObject() Object
	ToNull() Null
	ToBool() Bool
	ToString() String
	ToNumber() Number
	ToArray() Array
	ToMap() Map

	Accept(visitor Visitor) int
}

type Visitor interface {
	VisitNull(Null) int
	VisitBool(Bool) int
	VisitString(String) int
	VisitNumber(Number) int
	EnterArray(Array) int
	VisitArrayItem(index int, value Object) int
	LeaveArray(Array) int
	EnterMap(Map) int
	VisitMapItem(index int, key string, value Object) int
	LeaveMap(Map) int
}

type Map interface {
	Object
	Map(key string) Object
	Len() int
	Foreach(f func(index int, key string, value Object) int)
}

type Array interface {
	Object
	Index(index int) Object
	Len() int
	Foreach(f func(index int, value Object) int)
}

type Null interface {
	Object
	String() string
}

type Bool interface {
	Object
	Bool() bool
	String() string
}

type String interface {
	Object
	String() string
}

type Number interface {
	Object
	Int() int
	Int64() int64
	Uint64() uint64
	Float32() float32
	Float64() float64
	Number() interface{}
	String() string
}

//------------------------------------------------------------------

type jsonObjectImpl struct {
	impl Object
}

func (o *jsonObjectImpl) Type() uint32 {
	return o.impl.Type()
}

func (o *jsonObjectImpl) ToObject() Object {
	return o.impl
}

func (o *jsonObjectImpl) ToNull() Null {
	return nil
}

func (o *jsonObjectImpl) ToBool() Bool {
	return nil
}

func (o *jsonObjectImpl) ToString() String {
	return nil
}

func (o *jsonObjectImpl) ToNumber() Number {
	return nil
}

func (o *jsonObjectImpl) ToArray() Array {
	return nil
}

func (o *jsonObjectImpl) ToMap() Map {
	return nil
}

//func (o *jsonObjectImpl) Accept(visitor Visitor) int {
//	return o.impl.Accept(visitor)
//}

//------------------------------------------------------------------

type jsonMapImpl struct {
	jsonObjectImpl
	valuelist *list.List
	valuesmap map[string]*list.Element
}

type jsonPair struct {
	key   string
	value Object
}

func (m *jsonMapImpl) Type() uint32 {
	return TYPE_MAP
}

func (m *jsonMapImpl) Map(key string) Object {
	v, ok := m.valuesmap[key]
	if !ok {
		return nil
	}

	return v.Value.(*jsonPair).value
}

func (m *jsonMapImpl) Len() int {
	return len(m.valuesmap)
}

func (m *jsonMapImpl) ToMap() Map {
	return m
}

func (m *jsonMapImpl) Accept(visitor Visitor) int {
	r := visitor.EnterMap(m)
	if 0 != r {
		return r
	}

	index := -1
	for elem := m.valuelist.Front(); nil != elem; elem = elem.Next() {
		index++
		r := visitor.VisitMapItem(index, elem.Value.(*jsonPair).key, elem.Value.(*jsonPair).value)
		if 0 != r {
			return r
		}
	}

	return visitor.LeaveMap(m)
}

func (m *jsonMapImpl) Foreach(f func(index int, key string, value Object) int) {
	index := -1
	for elem := m.valuelist.Front(); nil != elem; elem = elem.Next() {
		index++
		r := f(index, elem.Value.(*jsonPair).key, elem.Value.(*jsonPair).value)
		if 0 != r {
			return
		}
	}
}

//------------------------------------------------------------------

type jsonArrayImpl struct {
	jsonObjectImpl
	values []Object
}

func (a *jsonArrayImpl) Type() uint32 {
	return TYPE_ARRAY
}

func (a *jsonArrayImpl) Index(index int) Object {
	if index >= len(a.values) {
		return nil
	}

	return a.values[index]
}

func (a *jsonArrayImpl) Len() int {
	return len(a.values)
}

func (a *jsonArrayImpl) ToArray() Array {
	return a
}

func (a *jsonArrayImpl) Accept(visitor Visitor) int {
	r := visitor.EnterArray(a)
	if 0 != r {
		return r
	}

	for index, v := range a.values {
		r := visitor.VisitArrayItem(index, v)
		if 0 != r {
			return r
		}
	}

	return visitor.LeaveArray(a)
}

func (a *jsonArrayImpl) Foreach(f func(index int, value Object) int) {
	for index, v := range a.values {
		r := f(index, v)
		if 0 != r {
			return
		}
	}
}

//------------------------------------------------------------------

type jsonNullImpl struct {
	jsonObjectImpl
}

func (n *jsonNullImpl) Type() uint32 {
	return TYPE_NULL
}

func (n *jsonNullImpl) Null() interface{} {
	return nil
}

func (n *jsonNullImpl) ToNull() Null {
	return n
}

func (n *jsonNullImpl) String() string {
	return "null"
}

func (a *jsonNullImpl) Accept(visitor Visitor) int {
	return visitor.VisitNull(a)
}

//------------------------------------------------------------------

type jsonBoolImpl struct {
	jsonObjectImpl
	value bool
}

func (b *jsonBoolImpl) Type() uint32 {
	return TYPE_BOOL
}

func (b *jsonBoolImpl) Bool() bool {
	return b.value
}

func (b *jsonBoolImpl) ToBool() Bool {
	return b
}

func (b *jsonBoolImpl) Accept(visitor Visitor) int {
	return visitor.VisitBool(b)
}

func (b *jsonBoolImpl) String() string {
	if b.value {
		return "true"
	}

	return "false"
}

//------------------------------------------------------------------

type jsonStringImpl struct {
	jsonObjectImpl
	value string
}

func (s *jsonStringImpl) Type() uint32 {
	return TYPE_STRING
}

func (s *jsonStringImpl) String() string {
	return s.value
}

func (s *jsonStringImpl) ToString() String {
	return s
}

func (s *jsonStringImpl) Accept(visitor Visitor) int {
	return visitor.VisitString(s)
}

//------------------------------------------------------------------

type jsonNumberImpl struct {
	jsonObjectImpl
	value float64
}

func (u *jsonNumberImpl) Type() uint32 {
	return TYPE_NUMBER
}

func (u *jsonNumberImpl) Int() int {
	return int(u.value)
}

func (u *jsonNumberImpl) Int64() int64 {
	return int64(u.value)
}

func (u *jsonNumberImpl) Uint64() uint64 {
	return uint64(u.value)
}

func (u *jsonNumberImpl) Float32() float32 {
	return float32(u.value)
}

func (u *jsonNumberImpl) Float64() float64 {
	return float64(u.value)
}

func (u *jsonNumberImpl) ToNumber() Number {
	return u
}

func (u *jsonNumberImpl) Number() interface{} {
	//	检查是否有小数部分:
	i64 := int64(u.value)
	postfix := u.value - float64(i64)
	if postfix != 0 {
		return u.value
	}

	return i64
}

func (u *jsonNumberImpl) Accept(visitor Visitor) int {
	return visitor.VisitNumber(u)
}

func (u *jsonNumberImpl) String() string {
	return fmt.Sprintf("%v", u.Number())
}

//------------------------------------------------------------------

func readMap(d *json.Decoder) (*jsonMapImpl, error) {
	m := new(jsonMapImpl)
	m.impl = m
	m.valuelist = list.New()
	m.valuesmap = make(map[string]*list.Element)

	for d.More() {
		t, err := d.Token()
		if nil != err {
			return nil, err
		}

		k, ok := t.(string)
		if !ok {
			return nil, fmt.Errorf("Except string")
		}

		v, err := readObject(d)
		if nil != err {
			return nil, err
		}

		_, ok = m.valuesmap[k]
		if ok {
			return nil, fmt.Errorf("Item with the same key has been exist: %s", k)
		}

		m.valuesmap[k] = m.valuelist.PushBack(&jsonPair{k, v})
	}

	d.Token()
	return m, nil
}

func readArray(d *json.Decoder) (*jsonArrayImpl, error) {
	a := new(jsonArrayImpl)
	a.impl = a
	a.values = make([]Object, 0, 2)

	for d.More() {

		v, err := readObject(d)
		if nil != err {
			return nil, err
		}

		a.values = append(a.values, v)
	}

	d.Token()
	return a, nil
}

func readObject(d *json.Decoder) (Object, error) {
	t, err := d.Token()
	if nil != err {
		return nil, err
	}

	switch t.(type) {
	case json.Delim:
		c := t.(json.Delim)
		switch c {
		case '{':
			return readMap(d)
		case '[':
			return readArray(d)
		default:
			return nil, fmt.Errorf("Unexpect delim: %c", c)
		}

	case string:
		o := new(jsonStringImpl)
		o.impl = o
		o.value = t.(string)
		return o, nil

	case bool:
		o := new(jsonBoolImpl)
		o.impl = o
		o.value = t.(bool)
		return o, nil

	case float64:
		o := new(jsonNumberImpl)
		o.impl = o
		o.value = t.(float64)
		return o, nil

	case nil:
		o := new(jsonNullImpl)
		o.impl = o
		return o, nil

	default:
		return nil, fmt.Errorf("Unexpect token")
	}
}

func LoadObject(r io.Reader) (Object, error) {
	return readObject(json.NewDecoder(r))
}

type PrintOptions struct {
}

type jsonSimplePrinter struct {
	writer  io.Writer
	options PrintOptions
}

func (p *jsonSimplePrinter) VisitNull(v Null) int {
	p.writer.Write([]byte(v.String()))
	return 0
}

func (p *jsonSimplePrinter) VisitBool(v Bool) int {
	p.writer.Write([]byte(v.String()))
	return 0
}

func (p *jsonSimplePrinter) VisitString(v String) int {
	p.writer.Write([]byte(`"`))
	p.writer.Write([]byte(v.String()))
	p.writer.Write([]byte(`"`))
	return 0
}

func (p *jsonSimplePrinter) VisitNumber(v Number) int {
	p.writer.Write([]byte(v.String()))
	return 0
}

func (p *jsonSimplePrinter) EnterArray(Array) int {
	p.writer.Write([]byte("["))
	return 0
}

func (p *jsonSimplePrinter) VisitArrayItem(index int, value Object) int {
	if 0 != index {
		p.writer.Write([]byte(","))
	}

	value.Accept(p)
	return 0
}

func (p *jsonSimplePrinter) LeaveArray(Array) int {
	p.writer.Write([]byte("]"))
	return 0
}

func (p *jsonSimplePrinter) EnterMap(Map) int {
	p.writer.Write([]byte("{"))
	return 0
}

func (p *jsonSimplePrinter) VisitMapItem(index int, key string, value Object) int {
	if 0 != index {
		p.writer.Write([]byte(`,`))
	}

	p.writer.Write([]byte(`"`))
	p.writer.Write([]byte(key))
	p.writer.Write([]byte(`":`))
	value.Accept(p)
	return 0
}

func (p *jsonSimplePrinter) LeaveMap(Map) int {
	p.writer.Write([]byte("}"))
	return 0
}

func NewSimplePrinter(writer io.Writer, options PrintOptions) Visitor {
	return &jsonSimplePrinter{writer, options}
}

func Version() string {
	return "1.0.0"
}
