# Publicar no GitHub

## Método automático

1. Entre no GitHub.
2. Crie um repositório **público e vazio** chamado `DualSenseHUB`.
3. Não crie README, licença ou `.gitignore` pela tela do GitHub.
4. Extraia este pacote.
5. Execute `publish-to-github.cmd`.
6. Cole a URL HTTPS do repositório.
7. Após o envio do código, execute `release-v1.0.12.cmd`.
8. Abra a aba **Actions** e aguarde o workflow concluir.
9. Abra **Releases** para conferir `DualSenseHUB.exe` e `SHA256SUMS.txt`.

## Usando GitHub Desktop

1. Adicione a pasta extraída como repositório local.
2. Publique como repositório público.
3. Crie e envie a tag `v1.0.12` pelo terminal ou Git.

## Sem Git instalado

Use **Add file → Upload files** no repositório e envie todos os arquivos,
incluindo a pasta `.github`. Depois crie a tag `v1.0.12`.

## Segurança

Não coloque senha, token ou chave pessoal dentro dos scripts. A autenticação
deve ocorrer pelo navegador, Git Credential Manager ou GitHub Desktop.
