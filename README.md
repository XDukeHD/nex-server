## Nex Server 
O Nex Server é um serviço que irá rodar no seu computador em background, ele monitora o status da sua CPU, uso de memória, uso de disco e uso de rede e afins, ele envia essas informações para o [Nex Viewer](https://github.com/XDukeHD/NexViewer), onde você visualiza essas informaçõs em tempo real podendo controlar també algumas ações do seu pc.

### Por que eu criei o Nex Viewer?
Simples, eu queria uma solução que me permitisse ver o status do meu notebook em tempo real pelo meu tablet, pois eu costumo ter um tablet no meu setup então decidi criar algo para isso, e nesse ponto eu pensei literalmente em criar o meu app do jeito que eu quero, então criei o Nex Viewer.

### Funcionalidades
- Visualização do uso de CPU, RAM, Armazenamento.
- Controle de Media Player (Play, Pause, Próxima música, Música anterior).
- Visualização de informações do sistema (Nome do dispositivo, IP, Sistema Operacional).

### Avisos
> [!WARNING] 
> O Nex Server foi desenvolvido e pensado para ser usado em uma máquina usando o Linux [Fedora](https://www.fedoraproject.org/) e foi testado apenas nesse sistema operacional, em tese ele deve funcionar em outras distros linux porém não garanto seu funcionamento, e ele não tem suporte para Windows ou MacOS, então se você usa um desses sistemas operacionais, infelizmente o Nex Server não é para você. (Inclusive se você usa o Windows te aconselho a pensar na sua escolha de vida :P) 

### Instalação 
Para instalar o Nex Server, você tem três opções:

# Alternativa 1: Baixar a source do projeto e compilar manualmente
1. Clone o repositório do Nex Server:
```bash
git clone https://github.com/XDukeHD/nex-server.git
```
2. Navegue até o diretório do projeto:
```bash 
cd nex-server
```
3. Compile o projeto usando o Makefile:
```bash
make build
```
4. Após a compilado você pode rodar o Nex Server usando o comando:
```bash
./nex-server
```
# Alternativa 2: Baixar o binário pré-compilado
1. Acesse a seção de [Releases](https://github.com/XDukeHD/nex-server/releases) 
2. Baixe o arquivo `nex-server-x.x.x` (substitua `x.x.x` pela versão mais recente).
3. Dê permissão de execução ao arquivo baixado:
```bash
chmod +x nex-server-x.x.x
```
4. Execute o Nex Server:
```bash
./nex-server-x.x.x
```
# Alternativa 3: Usar o Script de instalação (recomendado)
> [!WARNING]
> O script de instalação é recomendado pois já cria o serviço do Nex Server para você, ou seja, ele irá iniciar o Nex Server automaticamente toda vez que você ligar o computador, além de facilitar a instalação e configuração do Nex Server.
1. Cole o comando abaixo no terminal para baixar e executar o script de instalação:
```bash
curl -sL https://raw.githubusercontent.com/XDukeHD/nex-server/main/install.sh | bash
```
2. Siga as instruções do script para concluir a instalação.
3. Após a instalação, o Nex Server será iniciado automaticamente. Você pode verificar o status do serviço usando o comando:
```bash
systemctl status nex-server
``` 

### Contribuição
Contribuições são bem-vindas! Se você deseja contribuir para o Nex Server, seja feliz e abra um pull request com suas melhorias ou correções de bugs.
