# Verificação da v1.0.12

## Verificações automatizadas executadas

- `gofmt` sem alterações pendentes.
- `go vet ./...` concluído sem erros.
- `go test ./...` concluído sem erros.
- Cross-compilation Windows x64 concluída.
- Executável validado como PE32+ Windows GUI x86-64.
- O recurso de ícone contém 16, 20, 24, 32, 48, 64, 128 e 256 pixels.
- Cada entrada do recurso de ícone foi comparada byte a byte com
  `assets/DualSenseHUB.ico`.

## Testes físicos informados durante o desenvolvimento

- Overlay pelo Bluetooth.
- Leitura da bateria sem fio.
- Indicador de carregamento.
- Botão PS e desligamento.
- Aviso de bateria crítica em 5%.
- Overlay fechando ao desligar o controle.
- Uso em jogo sem conflito aparente nos botões.
- Ícone correto no Gerenciador de Tarefas e Menu Iniciar.

A validação física principal foi feita com DualSense Edge. A compatibilidade com
o DualSense padrão está implementada no código, mas merece testes adicionais da
comunidade.
