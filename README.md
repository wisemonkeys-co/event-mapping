# OCS Event Commons
Lib que reune logica de mapeamento de atributos de eventos externos para montagem de eventos do billing.

## Desenvolvimento
Executar o `go mod edit -replace` no projeto de fronteira (file, event, api)
```sh
# go mod edit -replace github.com/wisemonkeys-co/event-mapping=<local_path>
$ go mod edit -replace github.com/wisemonkeys-co/event-mapping=../event-mapping
```
