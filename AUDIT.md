# Guia de auditoria

Este arquivo indica onde verificar os comportamentos mais importantes antes de
executar ou redistribuir o programa.

## O programa acessa a internet?

Não. O código de execução não importa pacotes de rede e não possui endpoints,
telemetria, login, anúncios ou atualização remota.

Para conferir:

```bash
go list -deps .
```

Procure por `net/http`, bibliotecas de analytics ou clientes externos. Eles não
fazem parte do executável.

## Instalação e inicialização automática

Em `main.go`, procure por:

- `installAndRelaunchIfNeeded`
- `installStartupEntry`
- `installStartMenuShortcut`

A instalação ocorre somente no perfil do usuário e não solicita privilégios de
administrador.

## Leitura do controle

Em `main.go`, procure por:

- `registerControllerRawInput`
- `handleRawInput`
- `handleControllerReport`

O programa registra Raw Input e filtra dispositivos Sony pelos IDs reconhecidos.

## Bateria Bluetooth

Em `bluetooth_windows.go`, procure por:

- `requestEnhancedBluetoothMode`
- `activateDualSenseEnhancedReports`
- `tryDualSenseFeatureReports`

O programa abre o HID com compartilhamento habilitado, executa
`HidD_GetFeature` e fecha o handle. Não cria controle virtual.

## Desligamento

Em `main.go`, procure por `triggerShutdown`. O comando executado é equivalente a:

```text
shutdown.exe /s /t 10 /c "Seu computador irá desligar em 10 segundos"
```

## Arquivos e Registro

O programa usa:

```text
%LOCALAPPDATA%\DualSenseHUB\DualSenseHUB.exe
%LOCALAPPDATA%\DualSenseHUB\DualSenseHUB-debug.log
%APPDATA%\Microsoft\Windows\Start Menu\Programs\DualSenseHUB.lnk
HKCU\Software\Microsoft\Windows\CurrentVersion\Run\DualSenseHUB
```

Não grava em pastas do sistema e não cria serviço do Windows.
