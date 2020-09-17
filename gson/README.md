[TOC]

gson
===



## 目标：

简单易懂好用的json解析库。（**只用于解析，不用于编码。**）



## 设计原理：

`json.RawMessage`

官方文档说明：`RawMessage can be used to delay parsing part of a JSON message.`

(RawMessage可被用来延迟解析部分json消息。)

本库利用`json.RawMessage`延迟解析的特点，直到json值被用到时，再按需解析，并将解析结果缓存起来，避免二次解析带来的性能问题。



## 使用简介：

JSON对象结构：

   ```go
   type JSON struct {
     // contains filtered or unexported fields
   }
   ```

### 构造方式：

1. FromBytes

   ```go
   p := []byte(`{"a":123}`)
   js := gson.FromBytes(p)
   println(js.Key("a").Int())
   ```

2. FromString

   ```go
   s := `{"a":123}`
   js := gson.FromString(s)
   println(js.Key("a").Int())
   ```

3. FromReader

   ```go
   r := strings.NewReader(`{"a":123}`)
   js := gson.FromReader(r)
   println(js.Key("a").Int())
   ```

4. json.Unmarshal

   ```go
   var js gson.JSON
   json.Unmarshal([]byte(`{"a":123}`), &js)
   println(js.Key("a").Int())
   ```



### 原则说明

**本库所有函数、方法返回的`*JSON`对象均不为`nil`，方便链式操作。如果有错误信息，可通过`(*JSON).Err()`得到。**



### Object操作

#### Key

获取当前`Object`对象`key`的值，示例：

   ```go
   s := `{"a": 123}`
   println(gson.FromString(s).Key("a").Int()) // Output: 123
   ```

#### Keys

获取当前`Object`对象所有`key`，该操作可用于遍历`Object`，示例：

   ```go
   s := `{"a":123,"b":456}`
   println(gson.FromString(s).Keys()) // Output: [a b]
   
   s = `null`
   println(gson.FromString(s).Keys()) // Output: []
   println(gson.FromString(s).Err()) // Output: <nil>
   ```



### List操作

#### Index

获取数组中第`i`位的值，示例：

   ```go
   s := `[123, 456]`
   println(gson.FromString(s).Index(1).Int()) // Output: 456
   ```

#### Len

获取数组长度，示例：

   ```go
   s := `[123, 456]`
   println(gson.FromString(s).Len()) // Output: 2
   
   s = `null`
   println(gson.FromString(s).Len()) // Output: 0
   println(gson.FromString(s).Err()) // Output: <nil>
   ```



### 数值操作

- 支持取值`string`、`int`、`bool`(***暂不支持`float`类型，可通过`Value()`方法解析***)。示例：

   ```go
   s := `{"a":1}`
   js := gson.FromString(s).Key("a")
   println(js.Str()) // Output: 1
   println(js.Int()) // Output: 1
   println(js.Bool()) // Output: true
   ```

- `TryStr`、`TryInt`、`TryBool`取值时同时返回错误信息。

   ```go
   s := `{"a":"xyz"}`
   js := gson.FromString(s).Key("a")
   fmt.Println(js.TryStr())
   fmt.Println(js.TryInt())
   fmt.Println(js.TryBool())
   // Output:
   // xyz <nil>
   // 0 gson: path 'object.a' expected type: number, error: json: invalid number literal, trying to unmarshal "\"xyz\"" into Number
   // false gson: path 'object.a' expected type: bool, error: strconv.ParseBool: parsing "xyz": invalid syntax
   ```

- `Value`可用于自定义解析，也可用于解析`float`值，示例：

   ```go
   var f float64
   gson.FromString(`3.14`).Value(&f)
   println(f) // Output: +3.140000e+000
   ```

- **trick操作：`(*JSON).Str()`**

该方法不止可以解析`string`类型的值，还可以将原`json`串返回，使用时应注意，示例：

   ```go
   s := `{"a": {"b": 123}}`
   println(gson.FromString(s).Key("a").Str()) // Output: {"b": 123}
   ```



### 探测操作

`Type()`方法可以判断该`json`串是什么类型的，从而允许您一直探测下去，还原整个`json`结构：

```go
js := gson.FromString(`{"a":123}`)
switch js.Type() {
case gson.TypObject:
case gson.TypList:
case gson.TypString:
case gson.TypNumber:
case gson.TypBool:
default:
}
```



### “语法糖”

- Get

  `Get`方法允许您通过一个路径查找对应的值，示例：

  ```go
  s := `[[[],[{},{},{"a":[[[{"b":[123]}]]]}]]]`
  js := FromString(s)
  k := "[0][1][2].a[0][0][0].b[0]"
  println(js.Get(k).Int()) // Output: 123
  ```

  

- Any

  `Any`方法返回第一个出现的`key`，示例：

  ```go
  s := `{"errcode": 123}`
  println(FromString(s).Any("errcode", "err_code").Int()) // Outpu: 123
  
  s = `{"err_code": "123"}`
  println(FromString(s).Any("errcode", "err_code").Int()) // Outpu: 123
  ```

  

### 错误处理

错误类型有以下几种，返回错误详细信息：

```go
type WrappedError struct {
    Path     string   // 请求路径
    Expected GsonType // 期望类型
    Err      error    // 具体错误信息
}

type IndexOutOfRangeError struct {
    Path  string
    Index int // 请求index
    Range int // index允许范围
}

type KeyNotFoundError struct {
    Path string
    Key  string   // 请求key
    Keys []string // 所有key列表
}

type UnmarshalTypeError struct {
    Path     string
    Expected GsonType // 期望解析类型
    Real     GsonType // 实际类型
}
```

