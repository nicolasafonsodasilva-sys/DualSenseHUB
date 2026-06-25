# DualSenseHUB

Aplicativo leve e comunitário para Windows que mostra a bateria do controle
DualSense em um overlay e permite desligar o computador segurando o botão
**PS**.

> Projeto não oficial, sem vínculo, patrocínio ou aprovação da Sony Interactive Entertainment.

## Versão estável

**v1.0.12**

## Funcionalidades

- Mostra o overlay ao pressionar o botão **PS**.
- Lê a bateria por USB e Bluetooth.
- Mostra um raio quando o controle está carregando.
- Exibe **“Bateria fraca — conecte o carregador”** quando entra na faixa crítica representada como **5%**.
- Mudanças normais de carga, como 45% → 55%, atualizam silenciosamente e não abrem o overlay.
- Mostra o overlay ao conectar/desconectar o carregador e ao conectar o controle.
- Segurando **PS por 3 segundos**, agenda o desligamento do Windows em 10 segundos.
- Exibe a mensagem **“Seu computador irá desligar em 10 segundos”**.
- Instala uma cópia em `%LOCALAPPDATA%\DualSenseHUB\DualSenseHUB.exe`.
- Cria inicialização automática somente para o usuário atual.
- Cria um atalho `DualSenseHUB` no Menu Iniciar.
- Não aparece na barra de tarefas.
- Não usa rede, telemetria, anúncios, conta online ou servidor externo.

## Compatibilidade

- Windows 10/11 x64.
- DualSense padrão: PID `0CE6`.
- DualSense Edge: PID `0DF2`.
- USB e Bluetooth.

A validação física principal da v1.0.12 foi feita com um **DualSense Edge**.
O código também reconhece o DualSense padrão, mas testes adicionais da comunidade
são bem-vindos.

## Como a bateria Bluetooth funciona

No Bluetooth padrão, o pacote simples do DualSense contém os botões, mas não a
bateria. Quando esse pacote é detectado, o DualSenseHUB abre o dispositivo em
modo compartilhado, faz uma leitura de `feature report` e fecha o acesso. Essa
leitura faz o controle começar a enviar os relatórios Bluetooth completos `0x31`,
onde ficam a bateria e o estado de carregamento.

O programa não cria controle virtual, não remapeia botões e não controla luz,
vibração ou gatilhos. Mesmo assim, o relatório completo pode ser incompatível
com algum aplicativo antigo baseado em DirectInput. O modo permanece ativo até
o controle ser desligado e ligado novamente.

## Precisão da porcentagem

O DualSense informa níveis em faixas, não uma medição contínua de 1% em 1%.
Por isso, o overlay normalmente mostra valores como 5%, 15%, 25%, 35% e assim
por diante. O aviso em **5%** representa aproximadamente a faixa de **0–9%**.

## Instalação

Baixe `DualSenseHUB.exe` na seção **Releases** e execute uma vez. O programa
copia a versão instalada para:

```text
%LOCALAPPDATA%\DualSenseHUB\DualSenseHUB.exe
```

Depois que a cópia instalada iniciar, o arquivo baixado pode ser apagado.
Executar uma versão mais nova atualiza a cópia instalada.

### SmartScreen

O projeto é aberto, mas o executável não possui certificado comercial de
assinatura de código. O Windows pode mostrar “Editor desconhecido”. Confira o
hash da release e, para máxima confiança, compile a tag correspondente.

## Cancelar o desligamento

Durante os 10 segundos de espera, pressione `Win + R` e execute:

```bat
shutdown /a
```

## Auditoria rápida

Os comportamentos mais sensíveis estão concentrados nestes locais:

- `main.go`: instalação, inicialização automática, atalho, overlay e desligamento.
- `bluetooth_windows.go`: ativação dos relatórios completos do Bluetooth.
- `controller_report.go`: interpretação dos pacotes do controle.
- `battery_windows.go`: leitura auxiliar de bateria pelo Windows.
- `debuglog_windows.go`: log local de diagnóstico.

Consulte também `AUDIT.md` e `SECURITY.md`.

## Compilar

Requisitos:

- Go 1.23 ou compatível.
- Internet na primeira compilação para obter a ferramenta de build
  `go-winres v0.3.3`, usada apenas para incorporar o arquivo
  `assets/DualSenseHUB.ico` no executável.

PowerShell:

```powershell
.\build.ps1
```

Prompt de Comando:

```bat
build.cmd
```

O executável será criado em `dist\DualSenseHUB.exe`, acompanhado de
`SHA256SUMS.txt`.

O aplicativo em execução usa somente a biblioteca padrão do Go e APIs nativas
do Windows. A ferramenta `go-winres` participa apenas da compilação do recurso
de ícone.

## Build público

Os workflows em `.github/workflows/` executam formatação, `go vet`, testes,
compilação e geração do SHA-256. Uma tag como `v1.0.12` cria automaticamente uma
GitHub Release com o executável.

## Desinstalar

Execute `uninstall.cmd`. O script encerra o processo, remove a inicialização
automática, o atalho do Menu Iniciar e a pasta instalada.

## Licença

O código-fonte está sob a licença MIT. Consulte `LICENSE` e `NOTICE.md`.
