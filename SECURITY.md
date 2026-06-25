# Segurança e privacidade

## Escopo

DualSenseHUB não possui comunicação de rede, telemetria, anúncios, login ou
coleta de dados. O aplicativo não solicita privilégios de administrador.

## Comportamentos sensíveis e intencionais

O programa executa ações legítimas que também podem chamar atenção de análises
heurísticas:

- copia o próprio executável para LocalAppData;
- cria uma entrada de inicialização automática no registro do usuário;
- cria um atalho no Menu Iniciar;
- roda sem janela na barra de tarefas;
- lê dispositivos HID/Raw Input;
- ativa os relatórios Bluetooth completos do DualSense por uma leitura de feature report;
- pode chamar `shutdown.exe` depois de o botão PS permanecer pressionado por 3 segundos.

Esses comportamentos estão documentados e implementados no código aberto.

## Verificação recomendada

1. Baixe pela seção Releases do repositório oficial.
2. Compare o SHA-256 com `SHA256SUMS.txt`.
3. Confira o workflow público da tag.
4. Para máxima confiança, compile localmente a mesma tag.

## Bluetooth completo

A leitura de bateria sem fio exige que o DualSense transmita relatórios completos
`0x31`. O programa ativa esse estado por `HidD_GetFeature`. Alguns aplicativos
DirectInput antigos podem não aceitar esse formato. Para retornar ao modo padrão,
desligue e ligue novamente o controle.

## Relato de vulnerabilidade

Use o recurso privado de relato de vulnerabilidade do GitHub, quando habilitado.
Evite publicar detalhes exploráveis em uma issue antes de uma correção.
