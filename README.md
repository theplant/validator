Support define multiple validation rules for one struct.

This package is based on <https://github.com/go-playground/validator/>.

For this `User` struct:

```go
type User struct {
	Name    string
	Age     int
	Address Address
}

type Address struct {
	IP   string
	Area string
}
```

And fill it:

```go
user := User{
	Name: "I'm Name",
	Age:  15,
	Address: Address{
		IP:   "8.4.3",
		Area: "earth",
	},
}
```

We can define the validation rules:

```go
fullRules := []validator.Rule{
	{Field: "Name", Tag: "required,lte=20"},
	{Field: "Age", Tag: "min=20,max=100"},
	{Field: "Address.IP", Tag: "ip"},
}
```

If we `DoRules`:

```
validate := validator.New()
validate.DoRules(user, fullRules)
```

We can get:

```
validator.VErrors{
	{Field: "Age", Tag: "min", Param: "20", Message: ""},
	{Field: "Address.IP", Tag: "ip", Param: "", Message: ""},
}
```

We can also use other validation rules to check the `user`:

```go
addressRules := []validator.Rule{
	{Field: "Address.IP", Tag: "ipv6"},
	{Field: "Address.Area", Tag: "eq=Sun"},
}

validate.DoRules(user, addressRules)

// get:
//
// validator.VErrors{
//     {Field: "Address.IP", Tag: "ipv6", Param: "", Message: ""},
//     {Field: "Address.Area", Tag: "eq", Param: "Sun", Message: ""},
// }
```
