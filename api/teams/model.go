package teams

type Team struct {
  Index     string    `json:"index"`
  Name      string    `json:"name"`
}

func (t *Team) QueryArgs() []any {
     return []any{&t.Index, &t.Name}
}
