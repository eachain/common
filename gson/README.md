[TOC]

gson
===



## 目标：

简单易懂好用的json解析库。



## 设计原理：

`json.RawMessage`

官方文档说明：`RawMessage can be used to delay parsing part of a JSON message.`

(RawMessage可被用来延迟解析部分json消息。)

本库利用`json.RawMessage`延迟解析的特点，直到json值被用到时，再按需解析，并将解析结果缓存起来，避免二次解析带来的性能问题。



## 使用简介：

JSON对象结构：

   ```go
   type GSON struct {
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

4. json.Unmarshal

   ```go
   var js gson.GSON
   json.Unmarshal([]byte(`{"a":123}`), &js)
   println(js.Key("a").Int())
   ```



### 原则说明

**本库所有函数、方法返回的`JSON`对象均不为`nil`，方便链式操作。如果有错误信息，可通过`(JSON).Err()`得到。**



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

- 支持取值`string`、`int`、`float`、`bool`。示例：

   ```go
   s := `{"a":1}`
   js := gson.FromString(s).Key("a")
   println(js.Str()) // Output: 1
   println(js.Int()) // Output: 1
   println(js.Float()) // Output: 1
   println(js.Bool()) // Output: true
   ```

- **trick操作：`(*GSON).Str()`**

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
	if js.IsInt() {
		js.Int()
	} else {
		js.Float()
	}
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
type KeyNotFoundErr struct {
    Key  string
    Path string
}

type TypOpErr struct {
	Op   string
	Typ  Type
	Path string
}
```


### 修改

```go
js := gson.FromBytes(nil)
js.Get("a").Set("123")
println(js.Str()) // Output: {"a":"123"}
```

```go
js := gson.FromString(`{"a":123}`)
js.Get("a").Set(456)
println(js.Str()) // Output: {"a":456}
```

```go
js := gson.FromString(`{"a":1,"b":"2"}`)
js.Get("a").Remove()
println(js.Str()) // Output: {"b":"2"}
```


