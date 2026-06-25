# Recursos visuais

- `000.png` até `100.png`: estados do overlay de bateria.
- `charging_bolt.png`: indicador de carregamento.
- `low_battery.png`: aviso de bateria crítica.
- `DualSenseHUB.ico`: ícone multitamanho oficial aprovado para a v1.0.12.

Os arquivos PNG são incorporados ao executável por `//go:embed assets/*.png`.
O arquivo `.ico` é incorporado como recurso do Windows durante a compilação por
`go-winres`.

Consulte `NOTICE.md` antes de reutilizar ou redistribuir os recursos visuais.
