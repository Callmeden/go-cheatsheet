package main

import (
	"bufio"
	"fmt"
	"maps"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Printable interface {
	ToString() string
}

type ObjectValue struct {
	Data map[string]interface{}
}

type ListValue struct {
	Data []interface{}
} 

func (o ObjectValue) ToString() string {
	keys := make([]string, 0, len(o.Data))
	for k := range o.Data {
    	keys = append(keys, k)
	}
	sort.Strings(keys)
	
	sb := new(strings.Builder)
	sb.WriteString("{")
	for i, key := range keys {
		if i != 0 {
			sb.WriteString(",")
		}
		value := o.Data[key]
		formatted := formatValue(value)
		sb.WriteString(key);sb.WriteString(":");sb.WriteString(formatted) 
	}
	sb.WriteString("}")

	return sb.String()
}

func (l ListValue) ToString() string {
	sb := new(strings.Builder)
	sb.WriteString("[")
	for i, value := range l.Data {
		if i!=0 {
			sb.WriteString(",")
		}
		sb.WriteString(formatValue(value))
	}
	sb.WriteString("]")
	return sb.String()
}

type TypeBox struct {
	store map[string]interface{}
}

func formatValue(v interface{}) string {
    if p, ok := v.(Printable); ok {
        return p.ToString()
    }

	switch v := v.(type) {
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return strings.TrimSpace(v)
	default:
		return ""	
	}
}

func NewTypeBox() *TypeBox {
    return &TypeBox{store:make(map[string]interface{})}
}

func (tb *TypeBox) SetScalar(key, typ, raw string){
	switch typ {
	case "INT":
		value, err := strconv.Atoi(raw)
		if err != nil {
			return
		}
		tb.store[key] = value
	case "STRING":
		tb.store[key] = strings.TrimSpace(raw)
	case "FLOAT":
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return
		} 
		tb.store[key] = value
	}
}

func (tb *TypeBox) SetObject(key string, fields [][3]string){
	o := ObjectValue{Data: make(map[string]interface{})}
	for _, field := range fields{
		name, typ, value := field[0], field[1], field[2]
		switch typ{
		case "INT":
			ivalue, err := strconv.Atoi(value)
			if err != nil{
				return
			}
			o.Data[name] = ivalue
		case "FLOAT":
			fvalue, err := strconv.ParseFloat(value,64)
			if err != nil {
				return
			}
			o.Data[name] = fvalue
		case "STRING":
			o.Data[name] = strings.TrimSpace(value)	
		}
	}
	
	tb.store[key] = o
}


func (tb *TypeBox) PushValue(key, typ, raw string){
	oldValue, exists := tb.store[key]
	var data []interface{}
	var newValue interface{}

	switch typ {
		case "INT":
			value, err := strconv.Atoi(raw)
			if err != nil {
				return
			}
			newValue = value
			
		case "FLOAT":
			value, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				return
			}
			newValue = value	
		case "STRING":
			newValue = strings.TrimSpace(raw)
		default:
			return
	}

	if !exists {
		data = append(data, newValue)
		tb.store[key] = ListValue{Data:data}
	} else if listValue, ok := oldValue.(ListValue); ok {
		listValue.Data = append(listValue.Data, newValue)
		tb.store[key] = listValue
	} else {
		data = append(data, oldValue, newValue)
		tb.store[key] = ListValue{Data:data}
	}
}

func (tb *TypeBox) MergeObjects(target, source string){
	targetObj, exists := tb.store[target]
	
	if !exists {
		return
	} 

	sourceObj, exists := tb.store[source]
	if !exists {
		return
	}

	if sourceObj, ok := sourceObj.(ObjectValue); ok {
		if targetObj, ok2:= targetObj.(ObjectValue); ok2 {
			maps.Copy(targetObj.Data, sourceObj.Data) 
		}
	} 
}

func (tb *TypeBox) PrintKey(key string) string{
	data, exists := tb.store[key]
	if !exists {
		fmt.Println("null")
		return "null"
	}
	return formatValue(data)
}

func main() {
	tb := NewTypeBox()
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	q,err := strconv.Atoi(scanner.Text())
	if err!= nil {
		return
	}
	
	for i:= 0; i < q; i++ {
		if !scanner.Scan() {
    		break
		}
		line := scanner.Text()
		words := strings.Split(strings.TrimSpace(line), " ")
		switch words[0] {
		case "SET":
			key, typ, value := words[1], words[2], words[3]
			tb.SetScalar(key,typ,value)
		case "OBJECT":
			key := strings.TrimSpace(words[1])
			n,err := strconv.Atoi(words[2])
			if err!= nil {
				return
			}
			var fields [][3]string
			for j:= 0; j < n; j++ {
				if !scanner.Scan() {
    				break
				}
				newLine := scanner.Text()
				field := strings.Split(strings.TrimSpace(newLine), " ")
				fields = append(fields, [3]string(field))
			}
			tb.SetObject(key, fields)
		case "PUSH":
			key, typ, value := words[1], words[2], words[3]
			tb.PushValue(key,typ,value)
		case "MERGE":
			target, source := words[1], words[2]
			tb.MergeObjects(target, source)
		case "PRINT":
			key := words[1]
			fmt.Println(tb.PrintKey(key))	
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Ошибка сканирования:", err)
	}
}
