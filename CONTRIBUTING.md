# Como contribuir

1. Crie um fork do repositório.
2. Abra uma branch para a alteração.
3. Execute `gofmt -w .`.
4. Execute `go vet ./...`.
5. Execute `go test ./...`.
6. Compile com `build.ps1` ou `build.cmd`.
7. Abra um Pull Request explicando o comportamento alterado e como ele foi testado.

Não inclua executáveis compilados em commits comuns. Os binários oficiais devem
ser produzidos pelos workflows públicos do GitHub Actions.

Alterações no Bluetooth precisam explicar se enviam ou leem relatórios HID e se
mudam a forma como jogos enxergam o controle.
