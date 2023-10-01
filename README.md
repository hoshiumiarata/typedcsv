# typedcsv

This package provides functionality for reading and writing CSV files with structs.
It supports multiple tags to customize the behavior of the CSV reader and writer, as well as TextMarshaler/TextUnmarshaler interfaces.
Check out [documentation](https://pkg.go.dev/github.com/hoshiumiarata/typedcsv) of TypedCSVReader/TypedCSVWriter for more information.

## Examples

### Reading

#### CSV

```csv
name,year,repo,stars
rust,2010-07-07,https://github.com/rust-lang/rust,85700
go,2009-11-10,https://github.com/golang/go,115000
ruby,1995-01-01,https://github.com/ruby/ruby,20800
```

#### Struct

```go
type ProgrammingLanguage struct {
	Name        string    `csv:"name"`
	ReleaseDete time.Time `csv:"year" time_format:"2006-01-02"`
	Repo        string    `csv:"repo"`
	Stars       int       `csv:"stars"`
}
```

#### Read all

```go
file, err := os.Open("programming_languages.csv")
if err != nil {
    panic(err)
}
defer file.Close()

csvReader := typedcsv.NewReader[ProgrammingLanguage](csv.NewReader(file))
err = csvReader.ReadHeader()
if err != nil {
    panic(err)
}

languages, err := csvReader.ReadAll()
if err != nil {
    panic(err)
}

# languages[0] => ProgrammingLanguage{Name:"rust", ReleaseDete:time.Date(2010, time.July, 7, 0, 0, 0, 0, time.UTC), Repo:"https://github.com/rust-lang/rust", Stars:85700}
# languages[1] => ProgrammingLanguage{Name:"go", ReleaseDete:time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC), Repo:"https://github.com/golang/go", Stars:115000}
# languages[2] => ProgrammingLanguage{Name:"ruby", ReleaseDete:time.Date(1995, time.January, 1, 0, 0, 0, 0, time.UTC), Repo:"https://github.com/ruby/ruby", Stars:20800}
```

### Writing

#### Struct

```go
type ProgrammingLanguage struct {
	Name        string    `csv:"name"`
	ReleaseDete time.Time `csv:"year" time_format:"2006-01-02"`
	Repo        string    `csv:"repo"`
	Stars       int       `csv:"stars"`
}
```

#### Write

```go
file, err := os.Create("programming_languages.csv")
if err != nil {
    panic(err)
}
defer file.Close()

csvWriter := typedcsv.NewWriter[ProgrammingLanguage](csv.NewWriter(file))
err = csvWriter.WriteHeader()
if err != nil {
    panic(err)
}

languages := []ProgrammingLanguage{
    {Name: "rust", ReleaseDete: time.Date(2010, time.July, 7, 0, 0, 0, 0, time.UTC), Repo: "https://github.com/rust-lang/rust", Stars: 85700},
    {Name: "go", ReleaseDete: time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC), Repo: "https://github.com/golang/go", Stars: 115000},
    {Name: "ruby", ReleaseDete: time.Date(1995, time.January, 1, 0, 0, 0, 0, time.UTC), Repo: "https://github.com/ruby/ruby", Stars: 20800},
}

for _, language := range languages {
    err = csvWriter.WriteRecord(language)
    if err != nil {
        panic(err)
    }
}

csvWriter.Flush()
```

#### Result (CSV)

```csv
name,year,repo,stars
rust,2010-07-07,https://github.com/rust-lang/rust,85700
go,2009-11-10,https://github.com/golang/go,115000
ruby,1995-01-01,https://github.com/ruby/ruby,20800
```