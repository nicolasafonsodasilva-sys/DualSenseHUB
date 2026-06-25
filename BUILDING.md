# Compilação

## Requisitos

- Go 1.23 ou compatível.
- Windows 10/11 para executar o programa.
- Internet na primeira compilação para baixar `go-winres v0.3.3`, ferramenta
  usada somente para gerar o recurso de ícone do Windows.

## Windows — PowerShell

```powershell
Set-ExecutionPolicy -Scope Process Bypass
.\build.ps1
```

## Windows — Prompt de Comando

```bat
build.cmd
```

## Processo executado pelos scripts

```powershell
go run github.com/tc-hib/go-winres@v0.3.3 simply --arch amd64 --manifest none --icon assets/DualSenseHUB.ico
go fmt ./...
go vet ./...
go test ./...
go build -buildvcs=false -trimpath -ldflags "-H=windowsgui -s -w" -o dist\DualSenseHUB.exe .
```

`go-winres simply` carrega diretamente `assets/DualSenseHUB.ico` e o incorpora
como o recurso de ícone `#1`, sem redimensionar a arte do arquivo `.ico`.

## Linux — cross-compilation

```bash
go run github.com/tc-hib/go-winres@v0.3.3 simply --arch amd64 --manifest none --icon assets/DualSenseHUB.ico
mkdir -p dist
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
  go build -buildvcs=false -trimpath \
  -ldflags='-H windowsgui -s -w' \
  -o dist/DualSenseHUB.exe .
sha256sum dist/DualSenseHUB.exe > dist/SHA256SUMS.txt
```

## Testes

```bash
go test ./...
go vet ./...
```

Os testes cobrem a interpretação de relatórios USB, Bluetooth simples,
Bluetooth completo, estados de bateria e recursos do overlay.
