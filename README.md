# DualSenseHUB

Aplicativo leve e comunitário para Windows que exibe a bateria do controle **DualSense** em um overlay e permite desligar o computador usando o botão **PS**.

> [!IMPORTANT]
> Projeto não oficial, sem vínculo, patrocínio ou aprovação da Sony Interactive Entertainment.

## Vídeo de demonstração

Assista ao vídeo mostrando a instalação e o funcionamento do DualSenseHUB:

[▶ Assistir à demonstração do DualSenseHUB](https://youtu.be/D2w3TmOZ9SM)

## Versão estável

**v1.0.12**

## Funcionalidades

* Exibe o overlay ao pressionar o botão **PS**.

* Lê a bateria por USB e Bluetooth.

* Mostra um raio quando o controle está carregando.

* Exibe **“Bateria fraca — conecte o carregador”** quando a bateria está aproximadamente em **5%**.

* Atualiza silenciosamente mudanças normais de carga, como **45% → 55%**, sem abrir o overlay.

* Exibe o overlay ao:

  * conectar o controle;
  * conectar o carregador;
  * desconectar o carregador.

* Ao segurar o botão **PS por 3 segundos**, agenda o desligamento do Windows em 10 segundos.

* Exibe a mensagem **“Seu computador irá desligar em 10 segundos”**.

* Instala uma cópia em:

  ```text
  %LOCALAPPDATA%\DualSenseHUB\DualSenseHUB.exe
  ```

* Cria a inicialização automática somente para o usuário atual.

* Cria um atalho chamado `DualSenseHUB` no Menu Iniciar.

* Não aparece na barra de tarefas.

* Não usa rede, telemetria, anúncios, conta online ou servidor externo.

## Compatibilidade

* Windows 10 e Windows 11 x64.
* DualSense padrão — PID `0CE6`.
* DualSense Edge — PID `0DF2`.
* Conexão USB ou Bluetooth.

A validação física principal da versão **v1.0.12** foi realizada com um **DualSense Edge**.

O código também reconhece o DualSense padrão, mas testes adicionais da comunidade são bem-vindos.

## Limitação em tela cheia exclusiva

> [!WARNING]
> O overlay não é exibido sobre jogos executados em **tela cheia exclusiva**.

Essa limitação é intencional. Forçar a sobreposição nesse modo poderia causar incompatibilidade com sistemas anticheat ou fazer o aplicativo ser interpretado incorretamente por determinados jogos.

Para visualizar o overlay durante o jogo, use um destes modos:

* **Janela sem bordas**;
* **Tela cheia em janela**;
* **Janela normal**.

## Como a bateria Bluetooth funciona

No Bluetooth padrão, o pacote simples do DualSense contém os botões, mas não contém os dados da bateria.

Quando esse pacote é detectado, o DualSenseHUB:

1. Abre o dispositivo em modo compartilhado;
2. Faz uma leitura de `feature report`;
3. Fecha o acesso ao dispositivo.

Essa leitura faz o controle começar a enviar os relatórios Bluetooth completos `0x31`, que contêm a bateria e o estado de carregamento.

O programa:

* não cria um controle virtual;
* não remapeia botões;
* não controla luzes;
* não controla vibração;
* não controla gatilhos.

Mesmo assim, o relatório completo pode ser incompatível com algum aplicativo antigo baseado em DirectInput. Esse modo permanece ativo até o controle ser desligado e ligado novamente.

## Precisão da porcentagem

O DualSense informa a bateria em faixas, e não como uma medição contínua de 1% em 1%.

Por isso, o overlay normalmente exibe valores como:

```text
5%, 15%, 25%, 35%...
```

O aviso em **5%** representa aproximadamente a faixa entre **0% e 9%**.

## Instalação

1. Acesse a seção **Releases** do projeto.
2. Baixe o arquivo `DualSenseHUB.exe`.
3. Execute o arquivo uma vez.

O programa copiará a versão instalada para:

```text
%LOCALAPPDATA%\DualSenseHUB\DualSenseHUB.exe
```

Depois que a cópia instalada for iniciada, o arquivo baixado poderá ser apagado.

Ao executar uma versão mais nova, a cópia instalada será atualizada automaticamente.

## SmartScreen

O projeto possui código aberto, mas o executável não possui um certificado comercial de assinatura de código.

Por isso, o Windows pode exibir o aviso **“Editor desconhecido”**.

Para conferir a integridade do arquivo:

* verifique o hash disponibilizado na release;
* para máxima confiança, compile a tag correspondente diretamente pelo código-fonte.

## Cancelar o desligamento

Durante os 10 segundos de espera:

1. Pressione `Win + R`;
2. Execute o comando:

```bat
shutdown /a
```

## Auditoria rápida

Os comportamentos mais sensíveis estão concentrados nos seguintes arquivos:

| Arquivo                | Responsabilidade                                                     |
| ---------------------- | -------------------------------------------------------------------- |
| `main.go`              | Instalação, inicialização automática, atalho, overlay e desligamento |
| `bluetooth_windows.go` | Ativação dos relatórios completos do Bluetooth                       |
| `controller_report.go` | Interpretação dos pacotes do controle                                |
| `battery_windows.go`   | Leitura auxiliar da bateria pelo Windows                             |
| `debuglog_windows.go`  | Log local de diagnóstico                                             |

Consulte também:

* `AUDIT.md`
* `SECURITY.md`

## Compilar

### Requisitos

* Go 1.23 ou compatível;
* acesso à internet na primeira compilação para obter a ferramenta `go-winres v0.3.3`.

A ferramenta `go-winres` é usada somente para incorporar o arquivo `assets/DualSenseHUB.ico` ao executável.

### PowerShell

```powershell
.\build.ps1
```

### Prompt de Comando

```bat
build.cmd
```

O executável será criado em:

```text
dist\DualSenseHUB.exe
```

Também será gerado o arquivo:

```text
SHA256SUMS.txt
```

O aplicativo em execução utiliza somente a biblioteca padrão do Go e APIs nativas do Windows. A ferramenta `go-winres` participa apenas da compilação do recurso de ícone.

## Build público

Os workflows presentes em `.github/workflows/` executam:

* formatação;
* `go vet`;
* testes;
* compilação;
* geração do hash SHA-256.

Uma tag como `v1.0.12` cria automaticamente uma **GitHub Release** com o executável.

## Desinstalar

Execute:

```bat
uninstall.cmd
```

O script:

* encerra o processo;
* remove a inicialização automática;
* remove o atalho do Menu Iniciar;
* remove a pasta instalada.

## Licença

O código-fonte está disponível sob a licença MIT.

Consulte:

* `LICENSE`
* `NOTICE.md`
  ::: 
