# edit
Edit is an implementation of the Acme/Sam command language

# usage
```
ed, _ := text.Open(text.NewBuffer())
ed.Insert([]byte("Removing vowels isnt the best way to name things"), 0)

cmd, _ := edit.Compile("x,[aeiou],d")
cmd.Run(ed)

fmt.Printf("%s\n", ed.Bytes())

```
