**Reforma Tributária do**

**Consumo – Adequações NF-e / NFC-e**

**Nota Técnica 2025.002-RTC - Versão 1.**

**Março de 2025**


## NT 2025.002-RTC


## Reforma Tributária - Emenda Constitucional 132/

NT 2025.002-RTC

```
8.13. Evento: Manifestação sobre Pedido de Transferência de Crédito CBS em Operações de
Sucessão....................................................................................................................... 43
```
```
8.14. Evento: Manifestação do Fisco sobre Pedido de Transferência de Crédito de IBS em
Operações de Sucessão ............................................................................................... 44
```
```
8.15. Evento: Manifestação do Fisco sobre Pedido de Transferência de Crédito de CBS em
Operações de Sucessão ............................................................................................... 45
```
8.16. Evento: Cancelamento de Evento ................................................................................. 46

9. DANFE .................................................................................................................................. 47

ANEXO I - NCM DO IMPOSTO SELETIVO .................................................................................. 48

ANEXO II - CÓDIGO DE CLASSIFICAÇÃO TRIBUTÁRIA DO IMPOSTO SELETIVO .................. 48

ANEXO III - CÓDIGO DE CLASSIFICAÇÃO TRIBUTÁRIA DO IBS E DA CBS ............................ 48

ANEXO IV - CÓDIGO DE CLASSIFICAÇÃO DO CRÉDITO PRESUMIDO ................................... 48


## Reforma Tributária - Emenda Constitucional 132/

NT 2025.002-RTC

Controle de Versões

## Versão Publicação Descrição

## 2025.00 2 -

## RTC-v.1.

## 03/2025 Inserção de campos de controle e criação de eventos para utilização na apuração do IBS,

## CBS e IS

Histórico de Alterações / Cronograma

## Versão Histórico de atualizações Implantação

## Teste

## Implantação

## Produção

## 2025.00 2 -

## RTC-v.1.

## Inserção de campos de controle e criação de eventos para

## utilização na apuração do IBS, CBS e IS

## 01 /0 7 /2025 01 / 10 /


## Reforma Tributária - Emenda Constitucional 132/

NT 2025.002-RTC

1. Introdução

```
A Lei Complementar 214/2025 que institui o Imposto sobre Bens e Serviços (IBS), a Contribuição
Social sobre Bens e Serviços (CBS) e o Imposto Seletivo (IS), cria o Comitê Gestor do IBS e
altera a legislação tributária, definiu na Seção VIII – Disposições transitórias, Art. 62, a
obrigatoriedade para Estados, o Distrito Federal e os Municípios adaptarem os sistemas
autorizadores de Documentos Fiscais Eletrônicos (DFe) vigentes para utilização de leiaute
padronizado, que permita aos contribuintes informarem os dados relativos ao Imposto sobre Bens
e Serviços (IBS), Contribuição sobre Bens e Serviços (CBS) e Imposto Seletivo (IS).
```
```
Esta Nota Técnica substitui, no âmbito da NFe/NFCe, a RT NT 2024.002 - IBS/CBS v1.10, que
cria novos eventos e modifica o leiaute da NF-e e NFC-e, inserindo os grupos e campos opcionais
relacionados à tributação do Imposto sobre Bens e Serviços (IBS), da Contribuição sobre Bens e
Serviços (CBS) e do Imposto Seletivo (IS), em atendimento as alterações previstas na Emenda
Constitucional 132 de 20 de dezembro de 2023 e Lei Complementar 214 de 16 de janeiro de 2025
para implementação da Reforma Tributária, com data de implantação em ambiente de produção
prevista para outubro de 2025, de modo a viabilizar sua efetiva operacionalização a partir de
janeiro de 2026.
```
```
Vale destacar que, em Produção, no ano de 2025 as informações de tributação relativas ao IBS,
CBS e IS serão opcionais e não serão validadas. A partir de janeiro de 2026, as novas regras de
validação referentes a tributação do IBS e da CBS serão aplicadas.
```
```
Como as discussões envolvendo a implantação da Reforma Tributária ainda estão em curso,
esclarecemos que esta NT será ajustada ao longo do seu processo de execução, da mesma
forma como ocorre com as demais NT já implementadas.
```

## Reforma Tributária - Emenda Constitucional 132/

NT 2025.002-RTC

2. Tipos Básicos da Tributação

```
Em busca de uma padronização entre os diversos documentos fiscais eletrônicos existentes, esta
NT introduz o arquivo “DFeTiposBasicos_v1.00.xsd” ao conjunto dos arquivos que compõem do
schema de todos os Documentos Fiscais Eletrônicos - DF-e, entre eles a NF-e e NFC-e.
```
```
Este arquivo define de forma estruturada a previsão de campos a serem informados para o
registro das informações referentes a tributação do IBS e da CBS em um tipo complexo
referenciado no leiaute padrão da NF-e e NFC-e conforme estrutura demonstrada no item 5, e
também será utilizado nos demais documentos fiscais eletrônicos.
```
3. Código de Classificação Tributária do IBS/CBS

```
O grupo de informações do IBS, CBS e IS associado aos itens do documento fiscal contém o
Código de Situação Tributária (CST) e Código de Classificação Tributária (cClassTrib) do IBS,
CBS e IS.
```
```
O Informe Técnico RT 2024.001 divulga a publicação da tabela com esta codificação, que está
disponível no Portal Nacional da NF-e (www.nfe.fazenda.gov.br), na aba “Documentos”, opção
“Diversos”.
```
```
Cada código “cClassTrib” corresponde a um dispositivo específico da Lei Complementar 214 /
2025, tornando objetiva a informação do contribuinte sobre como é realizada a tributação do IBS e
da CBS para cada item da NF-e.
```
```
A tabela também contém indicadores que vinculam de forma dinâmica códigos “CST-IBS/CBS” e
“cClassTrib” com as Regras de Validação descritas na Nota Técnica 2025.002 – IBS/CBS/IS, ou
que contêm informações necessárias para a preparação das apurações assistidas do IBS e da
CBS, em atendimento ao disposto na Lei Complementar 214 / 2025.
```
```
Caso a versão final aprovada pelo Congresso Nacional altere algum dos dispositivos da Lei
Complementar 214 / 2025 a tabela será alterada para refletir esta modificação. Da mesma
maneira, a tabela poderá sofrer alterações em virtude de aperfeiçoamentos, novidades
introduzidas em sede de Regulamento, ou para atender a necessidades relacionadas com a
apuração assistida do IBS e da CBS.
```
4. Finalidade Débito e Finalidade Crédito da NF-e

```
Notas de Débito e Crédito são nomes de instrumentos utilizados mundialmente para documentar
situações contábeis onde é necessário corrigir informações comerciais que foram registradas em
um documento, que no Brasil é a Nota Fiscal.
```
```
Esta Nota Técnica cria na NF-e modelo 55 as finalidades de emissão correspondentes. O sentido
das palavras “débito” e “crédito” sempre se referem ao ponto de vista do emissor:
```
- Uma nota de débito documenta uma situação na qual o emitente registra um aumento no
    imposto devido (consequentemente, uma redução no imposto devido pelo adquirente, que é
    o destinatário);
- Uma nota de crédito documenta uma situação na qual o emitente registra uma redução no
    imposto devido (consequentemente, um aumento no imposto devido pelo adquirente, que é
    o destinatário);


## Reforma Tributária - Emenda Constitucional 132/

NT 2025.002-RTC

```
As finalidades de emissão “Nota de Ajuste” e “Nota Complementar”, já existentes, são casos
especiais de Nota de Débito; uma nota de entrada emitida para documentar, por exemplo, a
devolução de mercadoria que havia sido vendida a um consumidor final, é um caso especial de
Nota de Crédito.
```
```
A regulamentação do IBS disporá sobre a utilização de notas de crédito e notas de débito para
lançamentos de ajuste, com a finalidade de instrumentalizar a preparação da declaração assistida
a ser oferecida para os contribuintes, de maneira automatizada, a partir de documentos fiscais
eletrônicos, em cumprimento ao que preconiza a LC 214/2025. A menos que ocorra alteração na
regulamentação do ICMS e do IPI, notas de crédito e notas de débito não poderão ser utilizadas
para ajustes relativos a estes tributos.
```

## Reforma Tributária - Emenda Constitucional 132/

NT 2025.002-RTC

5. Protocolo da NFe

Alteração da seção 5.2.2 do MOC - Leiaute Mensagem de Retorno.

Retorno: Estrutura XML com o resultado do processamento da mensagem de envio de lote de NF-e.

Schema XML: retConsReciNFe_v4.00.xsd

* Para cada Protocolo de uma NF-e processada teremos o seguinte leiaute:

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
PR01 protNFe Raiz - - - - TAG raiz do Protocolo de recebimento da NFe
PR02 versao A PR01 N 1 - 1 2v2 Versão do leiaute das informações de Protocolo.
PR03 infProt G PR01 - 1 - 1 - Informações do Protocolo de resposta.
TAG a ser assinada
PR04 Id ID PR03 C 0 - 1 - Identificador da TAG a ser assinada, somente precisa ser informado se a UF assinar a resposta.
Em caso de assinatura da resposta pela SEFAZ preencher o campo com o Número do Protocolo,
precedido com o literal “ID”
PR05 tpAmb E PR03 N 1 - 1 1 Identificação do Ambiente: 1=Produção/2=Homologação
PR06 verAplic E PR03 C 1 - 1 1 - 20 Versão do Aplicativo que processou o Lote.
A versão deve ser iniciada com a sigla da UF nos casos de WS próprio ou a sigla SVAN ou SVRS nos
demais casos.
PR07 chNFe E PR03 N 1 - 1 44 Chave de Acesso da NF-e
PR08 dhRecbto E PR03 D 1 - 1 - Preenchido com a data e hora do processamento (informado também no caso de rejeição).
Formato: “AAAA-MM-DDThh:mm:ssTZD” (UTC – Universal Coordinated Time).
PR09 nProt E PR03 N 0 - 1 15 Número do Protocolo da NF-e, conforme item 4.3.5 do MOC
PR10 digVal E PR03 C 0 - 1 28 Digest Value da NF-e processada
Utilizado para conferir a integridade da NFe original.
PR11 cStat E PR03 N 1 - 1 4 Código do status da resposta
PR12 xMotivo E PR03 C 1 - 1 1 - 255 Descrição literal do status da resposta para a NF-e.
PR13 Sequência XML G PR03 0 - 1 Grupo de informações para envio de mensagens do interesse da SEFAZ (Criado na NT 2018.005)
PR14 cMsg E PR13 N 0 - 1 1 - 4 Código da Mensagem. (Criado na NT 2018.005)
PR15 xMsg E PR13 C 1 - 1 1 - 200 Mensagem da SEFAZ para o emissor. (Criado na NT 2018.005)
PR90 Signature G PR01 xml 0 - 1 - Assinatura XML do grupo identificado pelo atributo “Id”
A decisão de assinar a mensagem fica a critério da UF interessada.
```

## Reforma Tributária - Emenda Constitucional 132/

NT 2025.002-RTC

6. Leiaute da NF-e (Modelo 55 e 65)

```
Grupo B. Identificação da Nota Fiscal eletrônica
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
16 B12 cMunFG Código do Município de Ocorrência do Fato
Gerador do ICMS
```
```
E B01 N 1 - 1 7 Informar o município de ocorrência do fato gerador do ICMS.
Utilizar a Utilizar a Tabela de código de Município do IBGE
16a B12a cMunFGIBS Código do Município de consumo, fato
gerador do IBS / CBS
```
```
E B01 N 0 - 1 7 Informar o município de ocorrência do fato gerador do fato gerador
do IBS / CBS.
Campo preenchido somente quando “indPres = 5 (Operação
presencial, fora do estabelecimento)”, e não tiver endereço do
destinatário (Grupo: E05) ou local de entrega (Grupo: G01).
29 B25 finNFe Finalidade de emissão da NF-e E B01 N 1 - 1 1 1=NF-e normal;
2=NF-e complementar;
3=NF-e de ajuste;
4=Devolução de mercadoria.
5=Nota de crédito;
6=Nota de débito;
```
29. 1 B25.1 tpNFDebito Tipo de Nota de Débito CE B01 N 0 - 1 2 01=Transferência de créditos para Cooperativas;
    02=Anulação de Crédito por Saídas Imunes/Isentas;
    03=Débitos de notas fiscais não processadas na apuração;
    04=Multa e juros;
    05=Transferência de crédito de sucessão;
29.2 B25.2 tpNFCredito Tipo de Nota de Crédito CE B01 N 0 - 1 2 A definir.
29.3 B25a indFinal Indica operação com Consumidor final E B01 N 1 - 1 1 0=Normal;
    1=Consumidor final;
29.4 B25b indPres Indicador de presença do comprador no
    estabelecimento comercial no momento da
    operação

```
E B01 N 1 - 1 1 0=Não se aplica (por exemplo, Nota Fiscal complementar ou de
ajuste);
1=Operação presencial;
2=Operação não presencial, pela Internet;
3=Operação não presencial, Teleatendimento;
4=NFC-e em operação com entrega a domicílio;
5=Operação presencial, fora do estabelecimento; (incluído NT
2016/002)
9=Operação não presencial, outros.
29.5 B25c indIntermed Indicador de intermediador/marketplace E B01 N 0 - 1 1 0=Operação sem intermediador (em site ou plataforma própria)
1=Operação em site ou plataforma de terceiros
(intermediadores/marketplace)
```
- Considera-se intermediador/marketplace os prestadores de
    serviços e de negócios referentes às transações comerciais ou
    de prestação de serviços intermediadas, realizadas por pessoas


## Reforma Tributária - Emenda Constitucional 132/

NT 2025.002-RTC

```
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
jurídicas inscritas no Cadastro Nacional de Pessoa Jurídica -
CNPJ ou pessoas físicas inscritas no Cadastro de Pessoa
Física - CPF, ainda que não inscritas no cadastro de
contribuintes do ICMS.
```
- Considera-se site/plataforma própria as vendas que não foram
    intermediadas (por marketplace), como venda em site próprio,
    teleatendimento.
(Criado na NT 2020.006)
... ... ... ... ... ... ... ... ... ...
29f B31 gCompraGov Grupo de Compra Governamental G B01 0 - 1
29f.1 B32 tpCompraGov Tipo de compra governamental E B31 N 1 - 1 1 1=União
2=Estado
3=Distrito Federal
4=Município
29f.2 B33 pRedutor Percentual de redução de alíquota em
compra governamental

```
E B31 N 1 - 1 3v2- 4 Conforme o art. 473 da LC 214/2024.
```
```
Grupo N01. ICMS Normal e ST
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
164 N01 ICMS Informações do ICMS da Operação própria e ST CG M01 - 0 - 1 - Informar apenas um dos grupos de tributação do
ICMS (ICMS00, ICMS10, ...) (v2.0)
```
```
Grupo UB. Informações dos tributos IBS / CBS e Imposto Seletivo
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
324.0 1 UB0 1 IS Informações do Imposto Seletivo G H01 - 0 - 1 -
324.0 2 UB0 2 CSTIS Código de Situação Tributária do Imposto Seletivo E UB 01 N 1 - 1 3 Utilizar tabela CÓDIGO DE CLASSIFICAÇÃO
TRIBUTÁRIA DO IMPOSTO SELETIVO
324.0 3 UB0 3 cClassTribIS Código de Classificação Tributária do Imposto
Seletivo
```
```
E UB0 1 C 1 - 1 6 Utilizar tabela CÓDIGO DE CLASSIFICAÇÃO
TRIBUTÁRIA DO IMPOSTO SELETIVO
324.0 4 UB0 4 - x- Sequência XML G UB0 1 - 0 - 1 -
324.0 5 UB0 5 vBCIS Valor da Base de Cálculo do Imposto Seletivo E UB0 4 N 1 - 1 13v
324.0 6 UB0 6 pIS Alíquota do Imposto Seletivo E UB0 4 N 1 - 1 3v2- 4
324.0 7 UB0 7 pISEspec Alíquota específica por unidade de medida
apropriada
```
```
E UB0 4 N 0 - 1 3v2- 4
```
324. 08 UB 08 - x- Sequência XML G UB0 4 - 0 - 1 -
324 .09 UB09 uTrib Unidade de Medida Tributável E UB 08 C 1 - 1 1 - 6
324.10 UB10 qTrib Quantidade Tributável E UB 08 N 1 - 1 11v0- 4
324.11 UB11 vIS Valor do Imposto Seletivo E UB0 4 N 1 - 1 13v


## Reforma Tributária - Emenda Constitucional 132/

NT 2025.002-RTC

```
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
324 .12 UB 12 IBSCBS Informações do Imposto de Bens e Serviços -
IBS e da Contribuição de Bens e Serviços - CBS
```
### G H01 - 0 - 1 -

```
324.1 3 UB1 3 CST Código de Situação Tributária do IBS e CBS E UB 12 N 0 - 1 3 Utilizar tabela CÓDIGO DE CLASSIFICAÇÃO
TRIBUTÁRIA DO IBS E DA CBS
324.1 4 UB1 4 cClassTrib Código de Classificação Tributária do IBS e CBS E UB 12 C 0 - 1 6 Utilizar tabela CÓDIGO DE CLASSIFICAÇÃO
TRIBUTÁRIA DO IBS E DA CBS
324.15 UB15 gIBSCBS Grupo de Informações do IBS, CBS e Imposto
Seletivo
```
### CG UB 12 1 - 1

```
324.16 UB16 vBC Base de cálculo do IBS e CBS E UB15 N 1 - 1 3v2- 4
324.17 UB17 gIBSUF Grupo de Informações do IBS para a UF G UB15 - 1 - 1 -
324.18 UB18 pIBSUF Alíquota do IBS de competência das UF E UB17 N 1 - 1 3v2- 4 Alíquota vigente do IBS da UF
324.19 UB19 - x- Sequência XML G UB17 - 0 - 1 -
324.20 UB20 vTribOP Valor bruto do tributo na operação E UB19 N 1 - 1 13v2 Valor do tributo considerando BC x Alq do IBS,
sem considerar qualquer desoneração.
324.21 UB21 gDif Grupo de Informações do Diferimento G UB19 0 - 1
324.22 UB22 pDif Percentual do diferimento E UB21 N 1 - 1 3v2- 4
324.23 UB23 vDif Valor do Diferimento E UB21 N 1 - 1 13v
324.24 UB24 gDevTrib Grupo de Informações da devolução de tributos G UB19 0 - 1
324.25 UB25 vDevTrib Valor do tributo devolvido E UB24 N 1 - 1 13v2 Valor do tributo devolvido. No fornecimento de
energia elétrica, água, esgoto e gás natural e em
outras hipóteses definidas no regulamento
324.26 UB26 gRed Grupo de informações da redução da alíquota G UB17 - 0 - 1 -
324.27 UB27 pRedAliq Percentual da redução de alíquota E UB26 N 1 - 1 3v2- 4
324.28 UB28 pAliqEfet Alíquota Efetiva do IBS de competência das UF que
será aplicada a Base de Cálculo
```
```
E UB26 N 1 - 1 3v2- 4 Alíquota efetiva, após aplicação da redução de
alíquota.
324.29 UB29 gTribRegular Grupo de informações da Tributação Regular G UB17 - 0 - 1 - Grupo de informações da Tributação Regular,
caso não cumprida a condição
resolutória/suspensiva.
Exemplo 1: Art. 442, §4 da LC 214/2025.
Operações com ZFM e ALC.
Exemplo 2: Operações com suspensão do
tributo.
324.30 UB30 CSTReg Código de Situação Tributária do IBS e CBS E UB29 N 1 - 1 3 Informado como seria caso não cumprida a
condição resolutória/suspensiva.
Utilizar tabela CÓDIGO DE CLASSIFICAÇÃO
TRIBUTÁRIA DO IBS E DA CBS
324.31 UB31 cClassTribReg Código de Classificação Tributária do IBS e CBS E UB29 C 1 - 1 6 Informado como seria caso não cumprida a
condição resolutória/suspensiva.
Utilizar tabela CÓDIGO DE CLASSIFICAÇÃO
TRIBUTÁRIA DO IBS E DA CBS
```

## Reforma Tributária - Emenda Constitucional 132/

NT 2025.002-RTC

```
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
324.33 UB33 pAliqEfetReg Valor da alíquota E UB29 N 1 - 1 3v2- 4 Informado como seria caso não cumprida a
condição resolutória/suspensiva.
324.34 UB34 vTribReg Valor do Tributo (IBS) E UB29 N 1 - 1 13v2 Informado como seria caso não cumprida a
condição resolutória/suspensiva.
324.35 UB35 vIBSUF Valor do IBS de competência da UF E UB17 N 1 - 1 13v
324.36 UB36 gIBSMun Grupo de Informações do IBS para o município G UB15 - 1 - 1 -
324.37 UB37 pIBSMun Alíquota do IBS de competência do Município E UB36 N 1 - 1 3v2- 4 Alíquota vigente do IBS do Município
324.38 UB38 - x- Sequência XML G UB36 - 0 - 1 -
324.39 UB39 vTribOP Valor bruto do tributo na operação E UB38 N 1 - 1 13v2 Valor do tributo considerando BC x Alq do IBS,
sem considerar qualquer desoneração.
324.40 UB40 gDif Grupo de Informações do Diferimento G UB38 0 - 1
324.41 UB41 pDif Percentual do diferimento E UB40 N 1 - 1 3v2- 4
324.42 UB42 vDif Valor do Diferimento E UB40 N 1 - 1 13v
324.43 UB43 gDevTrib Grupo de Informações da devolução de tributos G UB38 0 - 1
324.44 UB44 vDevTrib Valor do tributo devolvido E UB43 N 1 - 1 13v2 Valor do tributo devolvido. No fornecimento de
energia elétrica, água, esgoto e gás natural e em
outras hipóteses definidas no regulamento
324.45 UB45 gRed Grupo de informações da redução da alíquota G UB36 - 0 - 1 -
324.46 UB46 pRedAliq Percentual da redução de alíquota E UB45 N 1 - 1 3v2- 4
324.47 UB47 pAliqEfet Alíquota Efetiva do IBS de competência do
Município que será aplicada a Base de Cálculo
```
```
E UB45 N 1 - 1 3v2- 4 Alíquota efetiva, após aplicação da redução de
alíquota.
324.48 UB48 gTribRegular Grupo de informações da Tributação Regular G UB17 - 0 - 1 - Grupo de informações da Tributação Regular,
caso não cumprida a condição
resolutória/suspensiva.
Exemplo 1: Art. 442, §4 da LC 214/2025.
Operações com ZFM e ALC.
Exemplo 2: Operações com suspensão do
tributo.
324.49 UB49 CSTReg Código de Situação Tributária do IBS e CBS E UB29 N 1 - 1 3 Informado como seria caso não cumprida a
condição resolutória/suspensiva.
Utilizar tabela CÓDIGO DE CLASSIFICAÇÃO
TRIBUTÁRIA DO IBS E DA CBS
324.50 UB50 cClassTribReg Código de Classificação Tributária do IBS e CBS E UB29 C 1 - 1 6 Informado como seria caso não cumprida a
condição resolutória/suspensiva.
Utilizar tabela CÓDIGO DE CLASSIFICAÇÃO
TRIBUTÁRIA DO IBS E DA CBS
324.52 UB52 pAliqEfetReg Valor da alíquota E UB29 N 1 - 1 3v2- 4 Informado como seria caso não cumprida a
condição resolutória/suspensiva.
324.53 UB53 vTribReg Valor do Tributo (IBS) E UB29 N 1 - 1 13v2 Informado como seria caso não cumprida a
condição resolutória/suspensiva.
```

## Reforma Tributária - Emenda Constitucional 132/

NT 2025.002-RTC

```
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
324.54 UB54 vIBSMun Valor do IBS de competência do Município E UB36 N 1 - 1 13v
324.55 UB55 gIBSCredPres Grupo de Informações do Crédito Presumido
referente ao IBS
```
```
G UB15 - 0 - 1 - Grupo de Informações do Crédito Presumido
do IBS, quando aproveitado pelo emitente do
documento. Exemplos:
1 - Aquisição de PR não contribuinte.
2 - Tomador de serviço de transporte de TAC
PF não contrib.
3 - Aquisição de pessoa física com destino a
reciclagem.
4 - Aquisição de bens móveis de PF não
contrib. para revenda (veículos / brechó).
5 - Regime opcional para cooperativa.
324.56 UB56 cCredPres Código de Classificação do Crédito Presumido E UB55 N 1 - 1 2 Utilizar tabela CÓDIGO DE CLASSIFICAÇÃO DO
CRÉDITO PRESUMIDO
324.57 UB57 pCredPres Percentual do Crédito Presumido E UB55 N 1 - 1 3v2- 4
324.58 UB58 vCredPres Valor do Crédito Presumido E UB55 CE 1 - 1 13v
324.59 UB59 vCredPresCondSus Valor do Crédito Presumido em condição
suspensiva.
```
```
E UB55 CE 1 - 1 13v2 Valor do Crédito Presumido Condição
Suspensiva. Preencher apenas para
cClassCredPres com indicação de Condição
Suspensiva.
324.60 UB60 gCBS Grupo de Informações da CBS G UB15 - 1 - 1 -
324.61 UB61 pCBS Alíquota da CBS E UB60 N 1 - 1 3v2- 4 Alíquota vigente da CBS.
324.62 UB62 - x- Sequência XML G UB60 - 0 - 1 -
324.63 UB63 vTribOp Valor bruto do tributo na operação E UB62 N 1 - 1 Valor do tributo considerando BC x Alq da CBS,
sem considerar qualquer desoneração.
324.64 UB64 gCBSCredPres Grupo de Informações do Crédito Presumido
referente a CBS
```
```
G UB62 - 0 - 1 Grupo de Informações do Crédito Presumido
da CBS, quando aproveitado pelo emitente do
documento. Exemplos:
1 - Aquisição de PR não contribuinte.
2 - Tomador de serviço de transporte de TAC
PF não contrib.
3 - Aquisição de pessoa física com destino a
reciclagem.
4 - Aquisição de bens móveis de PF não
contrib. para revenda (veículos / brechó).
5 - Regime opcional para cooperativa.
324.65 UB65 cCredPres Código de Classificação do Crédito Presumido E UB64 N 1 - 1 2 Utilizar tabela CÓDIGO DE CLASSIFICAÇÃO DO
CRÉDITO PRESUMIDO
324.66 UB66 pCredPres Percentual do Crédito Presumido E UB64 N 1 - 1 3v2- 4
324.67 UB67 vCredPres Valor do Crédito Presumido E UB64 CE 1 - 1 13v
```

## Reforma Tributária - Emenda Constitucional 132/

NT 2025.002-RTC

```
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
324.68 UB68 vCredPresCondSus Valor do Crédito Presumido em condição
suspensiva.
```
```
E UB64 CE 1 - 1 13v2 Valor do Crédito Presumido Condição
Suspensiva. Preencher apenas para
cClassCredPres com indicação de Condição
Suspensiva.
324.69 UB69 gDif Grupo de Informações do Diferimento G UB62 - 0 - 1
324.70 UB70 pDif Percentual do diferimento E UB69 N 1 - 1 3v2- 4
324.71 UB71 vDif Valor do Diferimento E UB69 N 1 - 1 13v
324.72 UB72 gDevTrib Grupo de Informações da devolução de tributos G UB62 - 0 - 1
324.73 UB73 vDevTrib Valor do tributo devolvido E UB72 N 1 - 1 13v2 Valor do tributo devolvido. No fornecimento de
energia elétrica, água, esgoto e gás natural e em
outras hipóteses definidas no regulamento
324.74 UB74 gRed Grupo de informações da redução da alíquota G UB60 - 0 - 1 -
324.75 UB75 pRedAliq Percentual da redução de alíquota E UB74 N 1 - 1 3v2- 4
324.76 UB76 pAliqEfet Alíquota Efetiva da CBS que será aplicada a Base
de Cálculo
```
```
E UB74 N 1 - 1 3v2- 4 Alíquota efetiva, após aplicação da redução de
alíquota.
324.77 UB77 gTribRegular Grupo de informações da Tributação Regular G UB17 - 0 - 1 - Grupo de informações da Tributação Regular,
caso não cumprida a condição
resolutória/suspensiva.
Exemplo 1: Art. 442, §4 da LC 214/2025.
Operações com ZFM e ALC.
Exemplo 2: Operações com suspensão do
tributo.
324.78 UB78 CSTReg Código de Situação Tributária do IBS e CBS E UB 77 N 1 - 1 3 Informado como seria caso não cumprida a
condição resolutória/suspensiva.
Utilizar tabela CÓDIGO DE CLASSIFICAÇÃO
TRIBUTÁRIA DO IBS E DA CBS
324.79 UB79 cClassTribReg Código de Classificação Tributária do IBS e CBS E UB 77 C 1 - 1 6 Informado como seria caso não cumprida a
condição resolutória/suspensiva.
Utilizar tabela CÓDIGO DE CLASSIFICAÇÃO
TRIBUTÁRIA DO IBS E DA CBS
324.81 UB81 pAliqEfetReg Valor da alíquota E UB 77 N 1 - 1 3v2- 4 Informado como seria caso não cumprida a
condição resolutória/suspensiva.
324.82 UB82 vTribReg Valor do Tributo (CBS) E UB 77 N 1 - 1 13v2 Informado como seria caso não cumprida a
condição resolutória/suspensiva.
324.83 UB83 vCBS Valor da CBS E UB60 N 1 - 1 13v
324.84 UB84 gIBSCBSMono Grupo de Informações do IBS e CBS em
operações com imposto monofásico
```
```
CG UB 12 1 - 1 Monofasia dos Combustíveis
```
```
324.85 UB85 qBCMono Quantidade tributada na monofasia E UB84 N 0 - 1 11v0- 4 Informar a BC quantidade conforme unidade de
medida estabelecida na legislação para o
produto.
```

## Reforma Tributária - Emenda Constitucional 132/

NT 2025.002-RTC

```
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
324.86 UB86 adRemIBS Alíquota ad rem do IBS E UB84 N 1 - 1 3v2- 4
324.87 UB87 adRemCBS Alíquota ad rem da CBS E UB84 N 1 - 1 3v2- 4
324.88 UB88 vIBSMono Valor do IBS monofásico E UB84 N 1 - 1 13v2 O valor do imposto é obtido pela multiplicação da
alíquota ad rem pela quantidade do produto
conforme unidade de medida estabelecida na
legislação.
324.89 UB89 vCBSMono Valor da CBS monofásica E UB84 N 1 - 1 13v2 O valor do imposto é obtido pela multiplicação da
alíquota ad rem pela quantidade do produto
conforme unidade de medida estabelecida na
legislação.
324.90 UB90 - x- Sequência XML G UB84 0 - 1 Uso em operações com combustíveis
derivados de petróleo (Gasolina A) [ou *Óleo
Diesel A*] para retenção do imposto sobre o
biocombustível a ser misturado. Art. 178 da
LC 214/2025.
324.91 UB91 qBCMonoReten Quantidade tributada sujeita à retenção na
monofasia
```
```
E UB90 N 0 - 1 11v0- 4 Informar a BC do ICMS sujeita a retenção em
quantidade conforme unidade de medida
estabelecida na legislação para o produto.
324.92 UB92 adRemIBSReten Alíquota ad rem do IBS sujeito a retenção E UB90 N 1 - 1 3v2- 4
324.93 UB93 vIBSMonoReten Valor do IBS monofásico sujeito a retenção E UB90 N 1 - 1 13v2 Valor do IBS com retenção, a ser somado ao
valor de IBS a ser recolhido.
324.93a UB93a adRemCBSReten Alíquota ad rem da CBS sujeito a retenção E UB90 N 1 - 1 3v2- 4
324.93b UB93b vCBSMonoReten Valor da CBS monofásica sujeita a retenção E UB90 N 1 - 1 13v2 Valor da CBS com retenção, a ser somado ao
valor de CBS a ser recolhido.
324.94 UB94 - x- Sequência XML G UB84 0 - 1 Tributação monofásica própria sobre
combustíveis cobrada anteriormente
324.95 UB95 qBCMonoRet Quantidade tributada retida anteriormente E UB94 N 1 - 1 11v0- 4 Informar a BC do IBS em quantidade conforme
unidade de medida estabelecida na legislação.
324.96 UB96 adRemIBSRet Alíquota ad rem do IBS retido anteriormente E UB94 N 1 - 1 3v2- 4 Alíquota ad rem do IBS, estabelecida na
legislação para o produto.
324.97 UB97 vIBSMonoRet Valor do IBS retido anteriormente E UB94 N 1 - 1 13v2 O valor do IBS é obtido pela multiplicação da
alíquota ad rem pela quantidade do produto
conforme unidade de medida estabelecida na
legislação.
324.98 UB98 adRemCBSRet Alíquota ad rem da CBS retida anteriormente E UB94 N 1 - 1 3v2- 4 Alíquota ad rem da CBS, estabelecida na
legislação para o produto.
324.98 UB98a vCBSMonoRet Valor da CBS retida anteriormente E UB94 N 1 - 1 13v2 O valor da CBS é obtido pela multiplicação da
alíquota ad rem pela quantidade do produto
conforme unidade de medida estabelecida na
legislação.
```

## Reforma Tributária - Emenda Constitucional 132/

NT 2025.002-RTC

```
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
324.99 UB99 - x- Sequência XML G UB84 0 - 1 Operações com diferimento, aplicado aos
biocombustíveis. Exemplo: operação do
produtor de biocombustível (usina).
324.100 UB100 pDifIBS Percentual do diferimento do imposto monofásico. E UB99 N 1 - 1 3v2- 4 A ser aplicado em vIBSMono.
324.101 UB101 vIBSMonoDif Valor do IBS mono diferido. E UB99 N 1 - 1 13v2 A ser deduzido do valor do IBS.
324.102 UB102 pDifCBS Percentual do diferimento do imposto monofásico E UB99 N 1 - 1 3v2- 4 A ser aplicado em vCBSMono
324.103 UB103 vCBSMonoDif Valor da CBS Mono diferido. E UB99 N 1 - 1 13v2 A ser deduzido do valor da CBS.
324.104 UB104 vTotIBSMonoItem Total de IBS Monofásico. E UB84 N 1 - 1 13v
324.105 UB105 vTotCBSMonoItem Total da CBS Monofásica. E UB84 N 1 - 1 13v
```
```
Grupo VB. Total do item da NF-e
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
325h VC01 vItem Valor Total do Item da NF-e E H01 N 1 - 1 13v2 Valor total do Item, correspondente à sua participação no total da nota.
A soma dos itens deverá corresponder ao total da nota.
```
```
Grupo VC. Referenciamento de item de outro Documento Fiscal Eletrônico - DF-e
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
325 i VC 01 DFeReferenciado Documento Fiscal Eletrônico Referenciado G H01 0 - 1 Grupo para referenciamento de itens de outro DF-e.
325 j VC 02 chaveAcesso Chave de acesso do DF-e referenciado E VC 01 N 1 - 1 44 Chave de acesso do DF-e referenciado.
325 k VC 03 nItem Número do item do documento referenciado.
E VC 01 N 1 - 1 3
```
```
Corresponde ao atributo “nItem” do elemento “det” do
documento original.
Se o documento referenciado não tiver item, indicar “1”
```
```
Grupo W03. Total da NF-e - IBS / CBS / IS
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
355.1 W31 IBSCBSSelTot Totais da NF-e com IBS, CBS e IS G A01 - 0 - 1 - O grupo de valores totais da NF-e deve ser
informado com o somatório do campo
correspondente dos itens.
```
```
O IBS, a CBS e o IS são por fora, por isso
seus valores devem ser adicionados ao valor
total da NF.
355.2 W32 gSel Grupo total do imposto seletivo G W31 - 0 - 1 -
355.3 W33 vBCIS Total da base de cálculo do imposto seletivo E W32 N 1 - 1 11v0- 4
355.4 W34 vIS Total do imposto seletivo E W32 N 1 - 1 13v
355.5 W35 vBCIBSCBS Valor total da BC do IBS e da CBS E W31 N 1 - 1 13v
355.6 W36 gIBS Grupo total do IBS G W31 - 1 - 1 -
355.7 W37 gIBSUFTot Grupo total do IBS da UF G W36 - 1 - 1 -
```

## NT 2025.002-RTC

- Reforma Tributária - Emenda Constitucional 132/
   - 1. Introdução Sumário
   - 2. Tipos Básicos da Tributação
   - 3. Código de Classificação Tributária do IBS/CBS
   - 4. Finalidade Débito e Finalidade Crédito da NF-e
   - 5. Protocolo da NFe
   - 6. Leiaute da NF-e (Modelo 55 e 65)
      - Grupo B. Identificação da Nota Fiscal eletrônica
      - Grupo N01. ICMS Normal e ST
      - Grupo UB. Informações dos tributos IBS / CBS e Imposto Seletivo
      - Grupo VB. Total do item da NF-e
      - Grupo VC. Referenciamento de item de outro Documento Fiscal Eletrônico - DF-e
      - Grupo W03. Total da NF-e - IBS / CBS / IS
   - 7. Regras de Validação
      - Grupo B. Identificação da Nota Fiscal eletrônica
      - Grupo BA. Documento Fiscal Referenciado
      - Grupo LA. Item / Combustível
      - Grupo UB. Informações dos tributos IBS / CBS e Imposto Seletivo
      - Grupo VB. Total do item da NF-e
      - Grupo VC. Referenciamento de item de outro Documento Fiscal Eletrônico - DF-e
      - Grupo W03. Total da NF-e - IBS / CBS / IS
      - Grupo 3A. Banco de Dados: NF-e Referenciada
   - 8. Eventos
      - 8.1. Lista de eventos
      - 8.2. Registro de Eventos
      - 8.3. Leiaute Mensagem de Retorno da Parte Geral
         - adquirente 8.4. Evento: Informação de efetivo pagamento integral para liberar crédito presumido do
      - 8.5. Evento: Solicitação de Apropriação de crédito presumido
      - 8.6. Evento: Destinação de item para consumo pessoal
      - 8.7. Evento: Perecimento, perda, roubo ou furto
      - 8.8. Evento: Aceite de débito na apuração por emissão de nota de crédito
      - 8.9. Evento: Imobilização de Item
      - 8.10. Evento: Solicitação de Apropriação de Crédito de Combustível
         - atividade do adquirente 8.11. Evento: Solicitação de Apropriação de Crédito para bens e serviços que dependem de
         - de Sucessão 8.12. Evento: Manifestação sobre Pedido de Transferência de Crédito de IBS em Operações
- Reforma Tributária - Emenda Constitucional 132/
   - 355.8 W38 vDif Valor total do diferimento E W37 N 1 - 1 13v # ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
   - 355.9 W39 vDevTrib Valor total de devolução de tributos E W37 N 1 - 1 13v
   - 355.11 W41 vIBSUF Valor total do IBS da UF E W37 N 1 - 1 13v
   - 355.13 W43 vDif Valor total do diferimento E W42 N 1 - 1 13v 355.12 W42 gIBSMunTot Grupo total do IBS do Município G W36 - 1 - 1 -
   - 355.14 W44 vDevTrib Valor total de devolução de tributos E W42 N 1 - 1 13v
   - 355.16 W46 vIBSMun Valor total do IBS do Município E W42 N 1 - 1 13v
   - 355.17 W47 vIBSTot Valor total do IBS E W36 N 1 - 1 13v
   - 355.18 W48 vCredPres Valor total do crédito presumido E W36 N 1 - 1 13v
   - 355.19 W49 vCredPresCondSus Valor total do crédito presumido em condição suspensiva. E W36 N 1 - 4 13v
   - 355.21 W51 vCredPres Valor total do crédito presumido E W50 N 1 - 1 13v 355.20 W50 gCBS Grupo total da CBS G W31 - 1 - 1 -
   - 355.22 W52 vCredPresCondSus Valor total do crédito presumido em condição suspensiva. E W50 N 1 - 4 13v
   - 355.23 W53 vDif Valor total do diferimento E W50 N 1 - 1 13v
   - 355.24 W54 vDevTrib Valor total de devolução de tributos E W50 N 1 - 1 13v
   - 355.26 W56 vCBS Valor total da CBS E W50 N 1 - 1 13v
   - 355.28 W58 vTotIBSMono Total do IBS monofásico E W57 N 1 - 1 13v 355.27 W57 gMono Grupo total da Monofasia G W31 - 0 - 1 -
   - 355.29 W59 vTotCBSMono Total da CBS monofásica E W57 N 1 - 1 13v
   - 355.30 W60 vTotNF Valor total da NF-e com IBS / CBS / IS E W57 N 1 - 1 13v


## Reforma Tributária - Emenda Constitucional 132/

## Reforma Tributária - Emenda Constitucional 132/

NT 2025.002-RTC

7. Regras de Validação

```
Grupo B. Identificação da Nota Fiscal eletrônica
Campo Modelo Regra de Validação Aplic. Msg Descrição Erro
B12a- 20 65 Se informado município do fato gerador IBS (tag: cMunUFIBS), validar se
operação é presencial fora do estabelecimento (tag: indPres = 5)
```
```
Obrig. 1000 Rejeição: Município do fato gerador do IBS deve ser informado
apenas em operação presencial fora do estabelecimento
B25- 60 55 Se NF-e complementar ou de crédito (tag: finNFe=2 ou 5):
```
- UF da NF-e referenciada diferente da UF do emitente (NF-e, NFC-e,
    NF modelo 1) (NT 2013/003)

```
Obrig. 678 Rejeição: NF referenciada com UF diferente da NF-e
complementar
```
```
B25- 80 55/65 Se finalidade da NF-e igual a crédito ou débito (tag:finNFe=5 ou 6):
Não pode ser informado ICMS (tag: ICMS), ISSQN (tag: ISSQN), IPI (tag:
IPI), II (tag: II), PIS (tag: PIS), PIS ST (tag: PISST), COFINS (tag:
COFINS), COFINS ST (tag: COFINSST), ICMS UF Destino (tag:
ICMSUFDest) e Imposto Devolvido (tag: impostoDevol).
```
```
Obrig. 1001 Rejeição: NF-e com finalidade de débito ou crédito somente
para IBS/CBS
```
```
B25- 90 55/65 Se finalidade da NF-e diferente de crédito ou débito (tag:finNFe=5 ou 6):
```
- Deve ser informado ICMS (tag: ICMS) ou ISSQN (tag: ISSQN).

```
Obrig. 1002 Rejeição: NF-e sem informação de ICMS / ISSQN
```
```
B25- 100 55/65 Se NF-e de crédito (tag: finNF-e = 5), somente pode ser referenciada NF-
e (tag: NFref/refNfe) e modelo deve ser 55.
```
```
Obrig. 1003 Rejeição: NF-e de crédito faz referência a documento fiscal
diferente de NF-e modelo 55
B25- 110 55 Se NF-e de devolução de mercadoria (tag: finNF-e = 4):
```
- Exige referenciamento de NF-e (tag: DFeReferenciado/refNFe) e do
    item da NF-e (tag: DFeReferenciado/nItem).

```
Obrig. 1102 Rejeição: NF-e de devolução de mercadoria exige
referenciamento do item da NF-e original
```
```
B25b- 50 65 NFC-e com entrega em domicílio (tag:indPres=4) e endereço do
destinatário não foi informado (id: E5, grupo: enderDest)
```
```
Obrig. 1004 Rejeição: NFC-e de entrega em domicílio e não informado
endereço do destinatário
B25b- 60 65 NFCe com operação presencial fora do estabelecimento (tag: indPres = 5)
e município do fato gerador do IBS não infomado (tag: cMunFGIBS)
```
```
Obrig. 1005 Rejeição: Operação presencial fora do estabelecimento deve
informar o município do fato gerador do IBS
B32- 10 55 Se informado tpCompraGov (nota de compra governamental), verificar se
alíquota de outros entes é diferente de zero:
```
- Se compra da união (tag: tpCompraGov = 1)
- Alíquota do IBS da UF (tag: pUBSUF) <> 0
- Alíquota do IBS do Município (tag: pUBSMun) <> 0
- Se compra estadual (tag: tpCompraGov = 2):
- Alíquota do IBS do Município (tag: pUBSMun) <> 0
- Alíquota da CBS (tag: pCBS) <> 0
- Se compra do DF (tag: tpCompraGov = 3):
- Alíquota da CBS (tag: pCBS) <> 0
- Se compra municipal (tag: tpCompraGov = 4):
- Alíquota do IBS da UF (tag: pUBSUF) <> 0
- Alíquota da CBS (tag: pCBS) <> 0
Observação: Conforme Art. 473, § 1o e seus Incisos I, II, III e IV da LC
    214/2025.

```
Obrig. 1008 Rejeição: Nota de compra governamental e alíquota dos
outros entes diferente de zero
```

## Reforma Tributária - Emenda Constitucional 132/

NT 2025.002-RTC

```
Campo Modelo Regra de Validação Aplic. Msg Descrição Erro
B34- 10 55 Se informado tipo de nota de crédito (tag: tpNFCredito), validar se
finalidade=5-Nota de crédito (tag:finNfe)
```
```
Obrig. 1009 Rejeição: Tipo de nota de crédito deve ser informado apenas
para nota com finalidade de nota de crédito
```
```
Grupo BA. Documento Fiscal Referenciado
Campo Modelo Regra de Validação Aplic. Msg Descrição Erro
BA01- 20 55/65 Não pode haver documento referenciado (tag: NFref), quando existe
referenciamento de item (tag: DFeReferenciado/nItem)
```
```
Obrig. 1010 Rejeição: NF-e com referenciamento de documento e de
item de documento
```
```
Grupo LA. Item / Combustível
Campo Modelo Regra de Validação Aplic. Msg Descrição Erro
LA01- 30 55/65 Não informado o grupo de combustível (tag: comb, id:LA01) e cClassTrib
vinculado a combustível (cClassTrib=410013 ou cClassTrib iniciado por
“620”)
```
```
Obrig. 1106 Rejeição: Não informado grupo de combustível para
cClassTrib de Combustível [nItem: 999]
```
```
Grupo UB. Informações dos tributos IBS / CBS e Imposto Seletivo
Campo Modelo Regra de Validação Aplic. Msg Descrição Erro
UB0 1 - 10 55/65 Não é permitido uso do Imposto Seletivo (grupo: gIS) para este
cClassTribIS.
Nota: Implementação Futura.
```
```
Obrig. 1011 Rejeição: Não é permitido uso do Imposto Seletivo para esta
classificação da operação
```
```
UB0 1 - 20 55/65 É exigido uso do Imposto Seletivo (grupo: gIS) para este cClassTrib.
Nota: Implementação Futura.
```
```
Obrig. 1012 Rejeição: É exigido o uso do Imposto Seletivo para esta
classificação da operação
UB0 1 - 30 55/65 É exigido uso do Imposto Seletivo (grupo: gIS) para este NCM.
Nota: Implementação Futura.
```
```
Obrig. 1013 Rejeição: É exigido o uso do Imposto Seletivo para esta
classificação da operação para este NCM
UB0 2 - 10 55/65 Se CST do Imposto Seletivo for informado, este deve existir na tabela de
Código de Situação Tributária (tag: gIS/CSTIS).
Nota: Implementação Futura.
```
```
Obrig. 1014 Rejeição: CST do Imposto Seletivo informado inexistente
```
```
UB0 3 - 10 55/65 Se cClassTribIS for informado, este deve existir na tabela de Classificação
Tributária do Imposto Seletivo (tag: gIS/cClassTribIS)
Nota: Implementação Futura.
```
```
Obrig. 1015 Rejeição: Classificação Tributária do Imposto Seletivo
informada inexistente
```
```
UB0 5 - 10 55/65 - Valor da Base de cálculo do Imposto Seletivo (vBCIS) deve ser igual ao
somatório de:
(+) vProd
(+) vServ
(+) vFrete
(+) vSeg
(+) vOutro
(+) vII
(-) vDesc
```
```
Obrig 1103 Rejeição: Valor da Base de cálculo do Imposto Seletivo
difere do somatório dos valores que a compõem
```

## Reforma Tributária - Emenda Constitucional 132/

NT 2025.002-RTC

```
Campo Modelo Regra de Validação Aplic. Msg Descrição Erro
(-) vPIS
(-) vCOFINS
(-) vICMS
(-) vICMSUFDest
(-) vFCP
(-) vFCPUFDest
(-) vICMSMono
(-) vISSQN
Exceção 1: Não subtrair o valor do PIS por Substituição Tributária
(PIST/vPIS) quando compor o valor total da NF-e (se
indSomaPISST=1);
Exceção 2: Não subtrair o valor do COFINS por Substituição Tributária
(COFINSST/vCOFINS) quando compor o valor total da NF-e (se
indSomaCOFINSST=1).
Nota: Implementação Futura.
UB0 6 - 10 55/65 Se cClassTribIS informado exigir grupo do Imposto Seletivo e exigir a
alíquota (tag: pIS) diferente de Zero, mas informada alíquota (tag: pIS) do
imposto seletivo igual a zero. (tag: gIS/pIS).
Nota: Implementação Futura.
```
```
Obrig. 1016 Rejeição: CST/cClassTribIS do Imposto Seletivo obriga
informação de alíquota de Imposto Seletivo
```
```
UB0 7 - 10 55 - 65 Se cClassTribIS informado exigir grupo do Imposto Seletivo e exigir a
alíquota específica (tag: gIS/pISEspec) diferente de Zero e NCM do item
for 2401, 2402, 2403, 2404, 2203, 2204, 2205, 2206, 2208.
Nota: Implementação Futura.
```
```
Obrig. 1017 Rejeição: Obrigatório informação de alíquota específica de
Imposto Seletivo
```
```
UB0 8 - 10 55/65 Se cClassTribIS informado exigir grupo do Imposto Seletivo (grupo:
seletivo) e unidade tributável (tag: gIS/uTrib) e quantidade tributável (tag:
gIS/qTrib) não informadas ou iguais a zero.
Nota: Implementação Futura.
```
```
Obrig. 1018 Rejeição: Unidade tributável (UTrib) e Quantidade tributável
(QTrib) do imposto seletivo não informados
```
```
UB11- 10 55/65 Se informado imposto seletivo (tag: imposto/seletivo):
Valor do IS (vIS) = BC (tag: vBCIS) * Alq (tag: pIS)
Observação 1: Se informada alíquota específica (tag: gIS/pISEspec):
Valor do IS (vIS) = qtd (tag: gIS/uTrib) * Alq (tag: gIS/pISEspec)
Observação 2: Aceitar uma tolerância de 0,01 a mais ou a menos.
Nota: Implementação Futura.
```
```
Obrig. 1019 Rejeição: Valor do Imposto Seletivo diferente de Base de
Cálculo x Alíquota
```
```
UB1 3 - 10 55/65 Se informada a tag CST do IBS/CBS (tag: IBSCBS/CST):
```
- CST inexistente

```
Obrig. 1020 Rejeição: CST do IBS/CBS informado inexistente
```
```
UB1 3 - 20 55/65 Se o CST do IBS/CBS (tag: IBSCBS/CST) informado não permitir a
informação do IBS/CBS:
```
- Informado indevidamente o grupo gIBSCBS (tag:
    imposto/IBSCBS/gIBSCBS) ou o grupo gIBSCBSMono (tag:
    imposto/IBSCBS/gIBSCBSMono)

```
Obrig. 1021 Rejeição: Grupo IBS/CBS não deve ser preenchido para o
CST informado
```
```
UB1 3 - 30 55/65 Se o CST do IBS/CBS (tag: IBSCBS/CST) informado: Obrig. 1022 Rejeição: Grupo IBS/CBS deve ser preenchido para o CST
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
Campo Modelo Regra de Validação Aplic. Msg Descrição Erro
```
- Grupo gIBSCBS (tag: imposto/IBSCBS/gIBSCBS) não informada informado
UB14- 10 55/65 Se cClassTrib (id: UB15, tag: IBSCBS/cClassTrib) for informado:
- cClassTrib inexistente;

```
Obrig. 1023 Rejeição: Classificação Tributária do IBS/CBS informada
inexistente
UB1 4 - 20 55/65 Se cClassTrib (id: UB15, tag: IBSCBS/cClassTrib) for informado:
```
- cClassTrib incompatível com CST (tag: IBSCBS/CST)

```
Obrig. 1024 Rejeição: Rejeição: Classificação Tributária incompatível
com o CST informado
UB16- 10 55/65 - Valor da Base de cálculo do IBS e CBS (gIBSCBS/vBC) deve ser igual ao
somatório de:
(+) vProd
(+) vServ
(+) vFrete
(+) vSeg
(+) vOutro
(+) vII
(-) vDesc
(-) vPIS
(-) vCOFINS
(-) vICMS
(-) vICMSUFDest
(-) vFCP
(-) vFCPUFDest
(-) vICMSMono
(-) vISSQN
(+) vIS
Exceção 1: Não subtrair o valor do PIS por Substituição Tributária
(PIST/vPIS) quando compor o valor total da NF-e (se
indSomaPISST=1);
Exceção 2: Não subtrair o valor do COFINS por Substituição Tributária
(COFINSST/vCOFINS) quando compor o valor total da NF-e (se
indSomaCOFINSST=1).
Nota: Implementação Futura.
```
```
Obrig 1104 Rejeição: Valor da Base de cálculo do IBS e CBS difere do
somatório dos valores que a compõem
```
```
UB18- 10 55/65 Alíquota do IBS da UF (tag: pIBSUF) deve ser igual a 0,1% para
documento com data de emissão no ano de 2026. Art. 343 da LC 214/25.
```
```
Obrig. 1026 Rejeição: Alíquota do IBS da UF deve ser igual a 0,1% para
documento emitido em 2026
UB18- 20 55/65 Alíquota do IBS da UF (tag: pIBSUF) deve ser igual a:
```
- 0,05% para documento com data de emissão nos anos de 2027 e
    2028. Art. 344 da LC 214/25.

```
Obrig. 1027 Rejeição: Alíquota do IBS da UF deve ser igual a 0,05% para
documento emitido em 2027 e 2028
```
```
UB20- 10 55/65 Se informado grupo IBS de competência das Unidades Federadas
(gIBSUF):
```
- Valor bruto do IBS das UF deve ser igual à multiplicação da Base de
    Cálculo pela Alíquota.
       vTribOP (tag: gIBSUF/vTribOP) = vBC (tag: gIBSCBS/vBC) * pIBSUF
       (tag: gIBSUF/pIBSUF)

```
Obrig. 1028 Rejeição: NF-e com valor bruto do tributo do IBS das UF
calculado incorretamente
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
Campo Modelo Regra de Validação Aplic. Msg Descrição Erro
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos.
UB22- 10 55/65 Não é permitido o uso de diferimento (grupo: gDif) para este cClassTrib Obrig. 1029 Rejeição: Não é permitido o uso de Diferimento para esta
classificação da operação
UB22- 20 55/65 É exigido o uso de diferimento da UF (grupo: gDif) para este cClassTrib Obrig. 1030 Rejeição: É exigido o uso de Diferimento da UF para esta
classificação da operação
UB23- 10 55/65 Se informado grupo do Diferimento (gIBSUF/gDif):
```
- Valor do Diferimento (vDif) deverá ser resultante da Base de Cálculo x
    Percentual do Diferimento (vBC x pDif)
Nota: Aceitar uma tolerância de 0,01 a mais ou a menos

```
Obrig. 1031 Rejeição: Valor do Diferimento da UF diferente de Base de
Cálculo x Percentual
```
```
UB26- 10 55/65 Não é permitido o uso de redução de alíquota (grupo: gRed) para este
cClassTrib.
```
```
Obrig. 1032 Rejeição: Não é permitido o uso de Redução de Alíquota
para esta classificação da operação
UB26- 20 55/65 É exigido o uso de redução de alíquota (grupo: gRed) para este
cClassTrib
```
```
Obrig. 1033 Rejeição: É exigido o uso de Redução de Alíquota para esta
classificação da operação
UB27- 10 55/65 Se informado grupo de Redução de Alíquota (gIBSUF/gRed):
```
- Percentual de Redução de Alíquota (pRedAliq) não é válido para este
    cClassTrib (GIBSCBS/cClassTrib)

```
Obrig. 1034 Rejeição: Percentual de redução de alíquota da UF não é
válido para este cClassTrib
```
```
UB28- 10 55/65 Se informado grupo de Redução de Alíquota (gIBSUF/gRed):
```
- Alíquota Efetiva (tag: pAliqEfet) deve ser o resultado da aplicação do
    percentual de redução da alíquota (tag: pRedAliq) na alíquota do IBS
    da UF (tag: pIBSUF).
Exemplo: Redução de 40% na alíquota:
    Alíquota vigente (A): 10%
    Redução na alíquota (R): 40%
    Alíquota Efetiva (E): E = A * (1 - R)
    E = 10 * (1 - 0,4) = 6

```
Obrig. 1035 Rejeição: Valor da Alíquota Efetiva do IBS da UF calculado
incorretamente
```
```
UB29- 10 55/65 cClassTrib do IBS/CBS exige informação do grupo de Tributação Regular
Estadual (gIBSUF /gTribReg)
```
```
Obrig. 1036 Rejeição: Classificação Tributária do IBS exige a informação
do grupo de Tributação Regular Estadual
UB29- 11 55/65 Se informado grupo da Tributação Regular Estadual (gIBSUF/gTribReg):
```
- Informado cClassTrib do IBS/CBS que não permite informação do
    grupo de Tributação Regular Estadual

```
Obrig. 1037 Rejeição: Classificação Tributária do IBS não permite a
informação do grupo de Tributação Regular Estadual
```
```
UB30- 10 55/65 Se informado o CST (gIBSUF/gTribReg/CSTReg) do grupo da Tributação
Regular Estadual:
```
- CST Regular (tag: gTribReg/CSTReg) inexistente

```
Obrig. 1038 Rejeição: CST Regular informado na Tributação Regular
Estadual inexistente
```
```
UB31- 10 55/65 Se informado grupo da Tributação Regular Estadual (gIBSUF/gTribReg):
```
- cClassTrib (tag: gTribReg/cClassTribReg) inexistente

```
Obrig. 1039 Rejeição: Classificação Tributária Regular informada na
Tributação Regular Estadual inexistente
UB34- 10 55/65 Se informado grupo da Tributação Regular Estadual (gIBSUF/gTribReg):
```
- Valor do Tributo Regular (vTribReg) deve ser resultante da Base de
    Cálculo x Alíquota Efetiva Regular (gTribReg/vBC x
    gTribReg/pAliqEfetReg)
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos

```
Obrig. 1040 Rejeição: Valor do Tributo Regular (vTribReg) do Estado
diferente de Base de Cálculo x Alíquota Efetiva Regular
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
Campo Modelo Regra de Validação Aplic. Msg Descrição Erro
UB35- 10 55/65 Se informado grupo IBS de competência das Unidades Federadas
(gIBSUF):
```
- Valor do IBS (vIBSUF) deverá ser resultante da Base de Cálculo x
    Alíquota (vBC [tag: gIBSCBS/vBC] x pIBSUF) - vDif – vDevTrib
Observação 1: Aceitar uma tolerância de 0,01 a mais ou a menos
Observação 2: Em caso de preenchimento do grupo de redução (pRed) a
    alíquota utilizada deverá ser a tag Alíquota Efetiva (pAliqEfet)
Observação 3: Conforme cClassTrib escolhido, o valor do crédito
    presumido (vCredPres) deve ser subtraído do total do IBS.

```
Obrig. 1041 Rejeição: NF-e com valor do IBS da UF calculado
incorretamente
```
```
UB37- 10 55/65 Alíquota do IBS do município (tag: pIBSMun) deve ser igual a 0,05% para
documento com data de emissão nos anos de 2027 e 2028. Art. 344 da
LC 214/2025.
```
```
Obrig. 1042 Rejeição: Alíquota do IBS da Município deve ser igual a
0,05% para documento emitido em 2027 e 2028
```
```
UB39- 10 55/65 Se informado grupo IBS de competência do Município (gIBSMun):
```
- Valor bruto do tributo do IBS do Município deve ser igual à
    multiplicação da Base de Cálculo pela Alíquota:
vTribOP (tag: gIBSMun/vTribOP) = vBC (tag: gIBSCBS/vBC) *
pIBSMun (tag: gIBSMun/pIBSMun)
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos.

```
Obrig. 1043 Rejeição: NF-e com valor bruto do tributo do IBS do
Município calculado incorretamente
```
```
UB40- 10 55/65 É exigido o uso de diferimento (grupo: gDif) Municipal para este
cClassTrib.
```
```
Obrig. 1044 Rejeição: É exigido o uso de Diferimento Municipal para esta
classificação da operação
UB42- 10 55/65 Se informado grupo do Diferimento (gIBSMun/gDif):
```
- Valor do Diferimento (vDif) deverá ser resultante da Base de Cálculo x
    Percentual do Diferimento (vBC x pDif)
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos.

```
Obrig. 1045 Rejeição: Valor do Diferimento do Município diferente de
Base de Cálculo x Percentual
```
```
UB46- 10 55/65 Se informado grupo de Redução de Alíquota (gIBSMun/gRed):
```
- Percentual de Redução de Alíquota (pRedAliq) não é válido para este
    cClassTrib (GIBSCBS/cClassTrib)

```
Obrig. 1046 Rejeição: Percentual de redução de alíquota do Município
não é válido para este cClassTrib
```
```
UB47- 10 55/65 Se informado grupo de Redução de Alíquota (gIBSMun/gRed):
```
- Valor da alíquota efetiva (pAliqEfet) deve ser igual a aplicação da
    redução da alíquota do IBS do Município (pRedAliq) na alíquota do
    IBS do Município (pIBSMun).
Exemplo: Redução de 40% na alíquota:
    Alíquota vigente (A): 10%
    Redução na alíquota (R): 40%
    Alíquota Efetiva (E): E = A * (1 - R)
    E = 10 * (1 - 0,4) = 6

```
Obrig. 1047 Rejeição: Valor da Alíquota Efetiva do IBS do Município
calculado incorretamente
```
```
UB48- 10 55/65 Código de Classificação Tributária do IBS exige informação do grupo de
Tributação Regular Municipal (gIBSMun/gTribReg)
```
```
Obrig. 1048 Rejeição: Classificação Tributária do IBS exige a informação
do grupo da Tributação Regular Municipal
UB48- 11 55/65 Se informado o grupo Tributação Regular Municipal (gIBSMun/gTribReg):
```
- Informado cClassTrib do IBS/CBS que não permite informação do
    grupo de Tributação Regular Municipal

```
Obrig. 1107 Rejeição: Classificação Tributária do IBS não permite a
informação do grupo de Tributação Regular Municipal
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
Campo Modelo Regra de Validação Aplic. Msg Descrição Erro
UB49- 10 55/65 Se informado o grupo Tributação Regular Municipal (gIBSMun/gTribReg):
```
- CST Regular informado deve existir na tabela de Classificação
    Tributária do IBS/CBS (tag: gTribReg/CSTReg)

```
Obrig. 1049 Rejeição: CST Regular informado na Tributação Regular
Municipal inexistente
```
```
UB50- 10 55/65 Se informado o grupo Tributação Regular Municipal (gIBSMun/gTribReg):
```
- cClassTribReg (tag: gTribReg/cClassTribReg) inexistente

```
Obrig. 1050 Rejeição: Classificação Tributária Regular informada na
Tributação Regular Municipal inexistente
UB53- 10 55/65 Se informado o grupo Tributação Regular Municipal (gIBSMun/gTribReg):
```
- Valor do Tributo Regular (vTribReg) deve ser resultante da Base de
    Cálculo x Alíquota Efetiva Regular (gTribReg/vBC x
    gTribReg/pAliqEfetReg)
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos.

```
Obrig. 1051 Rejeição: Valor do Tributo Regular (vTribReg) do Município
diferente de Base de Cálculo x Alíquota Efetiva Regular
```
```
UB54- 10 55/65 Se informado grupo IBS de competência do Município (gIBSMun):
```
- Valor do IBS (vIBSMun) deverá ser resultante da Base de Cálculo x
    Alíquota (vBC [tag: gIBSCBS/vBC] x pIBSMun) - vDif - vDevTrib
Observação 1: Aceitar uma tolerância de 0,01 a mais ou a menos.
Observação 2: Em caso de preenchimento do grupo de redução (pRed) a
    alíquota utilizada deverá ser a tag Alíquota Efetiva (pAliqEfet).
Observação 3: Conforme cClass escolhido o valor do crédito presumido
    (vCredPres) deve ser subtraído do total do IBS.

```
Obrig. 1052 Rejeição: NF-e com valor do IBS do Município calculado
incorretamente
```
```
UB55- 10 55/65 Não é permitido o uso de crédito presumido do IBS (grupo: gIBSCredPres)
para este código de classificação tributária (tag: gIBSCBS/cClassTrib).
```
```
Obrig. 1053 Rejeição: Não é permitido o uso de Crédito Presumido para
esta classificação da operação
UB55- 20 55/65 É exigido o uso de crédito presumido do IBS (grupo: gIBSCredPres) para
este código de classificação tributária (tag: GIBSCBS/cClassTrib).
```
```
Obrig. 1054 Rejeição: É exigido o uso de Crédito Presumido do IBS para
esta classificação da operação
UB58- 10 55/65 Se informado grupo do crédito presumido (tag: gIBSCBS/gIBSCredPres):
```
- Valor do Crédito Presumido (tag: vCredPres) deverá ser resultante da
    Base de Cálculo x Percentual do Crédito Presumido (vBC x
    pCredPres)
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos.

```
Obrig. 1055 Rejeição: Valor do Crédito Presumido do IBS diferente de
Base de Cálculo x Percentual
```
```
UB59- 10 55/65 Se informado Valor do Crédito Presumido em condição suspensiva
(tag:gIBSCredPres/vCredPresCondSus):
```
- Código de classificação do crédito presumido (tag:
    gIBSCredPres/cCredPres) <> “ 4 - Aquisição de bens móveis de PF não
    contrib. para revenda (veículos / brechó)”

```
Obrig. 1056 Rejeição: Valor do Crédito Presumido em condição
suspensiva deve ser informado somente para cCredPres = 4
```
- Aquisição de bens móveis de PF não contrib. para revenda
(veículos / brechó)

```
UB63- 10 55/65 Se informado grupo CBS (gCBS):
```
- Valor bruto da CBS deve ser igual à multiplicação da Base de Cálculo
    pela Alíquota.
       vTribOP (tag: gCBS/vTribOP) = vBC (tag: gIBSCBS/vBC) * pCBS
       (tag: gCBS/pCBS)
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos.

```
Obrig. 1057 Rejeição: NF-e com valor bruto do tributo da CBS calculado
incorretamente
```
```
UB64- 10 55/65 Se cClassTrib exige o uso de crédito presumido para a CBS (grupo:
gCBSCredPres):
```
- Crédito presumido para a CBS não informado

```
Obrig. 1058 Rejeição: Crédito presumido para a CBS não informado
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
Campo Modelo Regra de Validação Aplic. Msg Descrição Erro
UB67- 10 55/65 Se informado grupo do crédito presumido (gCBS/gCBSCredPres):
```
- Valor do Crédito Presumido (vCredPres) deverá ser resultante da
    Base de Cálculo x Percentual do Crédito Presumido (vBC x
    pCredPres)
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos.

```
Obrig. 1059 Rejeição: Valor do Crédito Presumido da CBS diferente de
Base de Cálculo x Percentual
```
```
UB68- 10 55/65 Se informado Valor do Crédito Presumido em condição suspensiva
(tag:gCBSCredPres/vCredPresCondSus):
```
- Código de classificação do crédito presumido (tag:
    gCBSCredPres/cCredPres) <> “ 4 - Aquisição de bens móveis de PF
    não contrib. para revenda (veículos / brechó)”

```
Obrig. 1060 Rejeição: Valor do Crédito Presumido em condição
suspensiva deve ser informado somente para cCredPres = 4
```
- Aquisição de bens móveis de PF não contrib. para revenda
(veículos / brechó)

```
UB69- 10 55/65 É exigido o uso de diferimento (grupo: gDif) da CBS para este cClassTrib. Obrig. 1061 Rejeição: É exigido o uso de Diferimento da CBS para esta
classificação da operação
UB71- 10 55/65 Se informado grupo do Diferimento (gCBS/gDif):
```
- Valor do Diferimento (vDif) deverá ser resultante da Base de Cálculo x
    Percentual do Diferimento (vBC x pDif)
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos.

```
Obrig. 1062 Rejeição: Valor do Diferimento da CBS diferente de Base de
Cálculo x Percentual
```
```
UB75- 10 55/65 Se informado grupo de Redução de Alíquota (gCBS/gRed):
```
- Percentual de Redução de Alíquota (pRedAliq) não é válido para este
    cClassTrib (IBSCBS/cClassTrib)

```
Obrig. 1063 Rejeição: Percentual de redução de alíquota da CBS não é
válido para este cClassTrib
```
```
UB76- 10 55/65 Se informado grupo de Redução de Alíquota (gCBS/gRed):
```
- Valor da alíquota efetiva (pAliqEfet) deve ser igual a aplicação da
    redução da alíquota da CBS (pRedAliq) na alíquota da CBS (pCBS).
Exemplo: Redução de 40% na alíquota:
    Alíquota vigente (A): 10%
    Redução na alíquota (R): 40%
    Alíquota Efetiva (E): E = A * (1 - R)
    E = 10 * (1 - 0,4) = 6

```
Obrig. 1064 Rejeição: Valor da Alíquota Efetiva da CBS calculado
incorretamente
```
```
UB77- 10 55/65 Código de Classificação Tributária exige informação do grupo de
Tributação Regular da CBS (gCBS/gTribReg)
```
```
Obrig. 1065 Rejeição: Classificação Tributária da CBS informada exige
informação da Tributação Regular da CBS
UB77- 11 55/65 Se informado grupo da Tributação Regular da CBS (gCBS/gTribReg):
```
- Informado cClassTrib do IBS/CBS que não permite informação do
    grupo de Tributação Regular da CBS

```
Obrig. XXXX Rejeição: Classificação Tributária da CBS não permite a
informação do grupo de Tributação Regular da CBS
```
```
UB78- 10 55/65 Se informado grupo da Tributação Regular da CBS (gCBS/gTribReg):
```
- CST Regular deve existir na tabela de Classificação Tributária do
    IBS/CBS (tag: gTribReg/CSTReg)

```
Obrig. 1066 Rejeição: CST Regular informado na Tributação Regular da
CBS inexistente
```
```
UB79- 10 55/65 Se informado grupo da Tributação Regular da CBS (gCBS/gTribReg):
```
- cClassTrib Regular (tag: gTribReg/cClassTribReg) inexistente

```
Obrig. 1067 Rejeição: Classificação Tributária Regular informada na
Tributação Regular da CBS inexistente
UB82- 10 55/65 Se informado grupo da Tributação Regular da CBS (gCBS/gTribReg):
```
- Valor do Tributo Regular (vTribReg) deve ser resultante da Base de
    Cálculo x Alíquota Efetiva Regular (gTribReg/vBC x

```
Obrig. 1068 Rejeição: Valor do Tributo Regular (vTribReg) da CBS
diferente de Base de Cálculo x Alíquota Efetiva Regular
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
Campo Modelo Regra de Validação Aplic. Msg Descrição Erro
gTribReg/pAliqEfetReg)
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos
UB83- 10 55/65 Se informado grupo CBS (gCBS):
```
- Valor da CBS (vCBS) deverá ser resultante da Base de Cálculo x
    Alíquota (vBC [tag: gIBSCBS/vBC] x pCBS) - vDif - vDevTrib
Observação 1: Aceitar uma tolerância de 0,01 a mais ou a menos.
Observação 2: Em caso de preenchimento do grupo de redução (pRed) a
    alíquota utilizada deverá ser a tag Alíquota Efetiva (pAliqEfet).
Observação 3: Conforme cClass escolhido o valor do crédito presumido
    (vCredPres) deve ser subtraído do total do IBS.

```
Obrig. 1069 Rejeição: NF-e com valor da CBS calculado incorretamente
```
```
UB94- 10 55/65 Se Tributação Monofásica de Combustível cobrada anteriormente
(cClassTrib=620003):
```
- Não informa a Tributação Monofásica Retida Anteriormente (id: UB94)
    própria sobre combustíveis cobrada anteriormente, observado o art.
    180 da LC 214 / 2025

```
Obrig. 1108 Rejeição: Não informa a Tributação Monofásica Retida
Anteriormente
```
```
UB94- 20 55/65 Se cClassTrib<>620003:
```
- Informada indevidamente a Tributação Monofásica Retida
    anteriormente (id: UB94) própria sobre combustíveis cobrada
    anteriormente, observado o art. 180 da LC 214/2025.

```
Obrig. 1109 Rejeição: Informada indevidamente a Tributação Monofásica
Retida anteriormente
```
```
UB104- 10 55/65 Se informado grupo do IBS e CBS monofásico (grupo: gIBSCBSMono):
```
- Valor total do IBS Monofásico (tag: vTotIBSMono) deverá ser
    resultante de: vIBSMono + vIBSMonoReten -vIBSMonoDif
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos.

```
Obrig. 1070 Rejeição: Valor do IBS monofásico calculado incorretamente
```
```
UB105- 10 55/65 Se informado grupo do IBS e CBS monofásico (grupo: gIBSCBSMono):
```
- O valor total da CBS Monofásica (tag: vTotCBSMono) deverá ser
    resultante de: vCBSMono + vCBSMonoReten -vCBSMonoDif
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos.

```
Obrig. 1071 Rejeição: Valor da CBS monofásico calculado
incorretamente
```
```
Grupo VB. Total do item da NF-e
Campo Modelo Regra de Validação Aplic. Msg Descrição Erro
VB 01 - 10 55/65 Se não é operação de Faturamento Direto para veículos novos (tpOp =
nulo ou tpOp <> 2, id:J02):
```
- Valor total do Item (vItem) deve ser igual ao somatório de:
    (+) vProd
    (-) vDesc
    (-) vICMSDeson, se indDeduzDeson=1
    (+) vICMSST
    (+) vICMSMonoReten
    (+) vFCPST
    (+) vFrete

```
Obrig. 1105 Rejeição: Valor total do Item (vItem) difere do somatório dos
valores que o compõem
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
Campo Modelo Regra de Validação Aplic. Msg Descrição Erro
(+) vSeg
(+) vOutro
(+) vII
(+) vIPI
(+) vIPIDevol
(+) vServ
(+) vPIS (id: R06, campo: PISST/vPIS), se indSomaPISST=1
(+) vCofins (id: T06, campo: COFINSST/vCOFINS), se
indSomaCOFINSST =1 (NT 2020.005)
(+) vIBSUF
(+) vIBSMun
(+) vCBS
(+) vIS
(+) vTotIBSMonoItem
(+) vTotCBSMonoItem
```
- Exceção 1 : Em 2026 não somar vIBSUF, vIBSMun, vCBS, vIS,
    vTotIBSMonoItem, vTotCBSMonoItem. * A confirmar.
- Observação 1: Implementação Futura.
VB 01 - 20 55 Se informada operação de Faturamento Direto para veículos novos (tpOp
= 2, id:J02):
- Valor total do Item (vItem) deve ser igual ao somatório de:
(+) vProd
(-) vDesc
(-) vICMSDeson, se indDeduzDeson=1
(+) vFrete
(+) vSeg
(+) vOutro
(+) vII
(+) vIPI
(+) vServ
(+) vPIS (id: R06, campo: PISST/vPIS), se indSomaPISST=1
(+) vCofins (id: T06, campo: COFINSST/vCOFINS), se
indSomaCOFINSST =1 (NT 2020.005)
(+) vIBSUF
(+) vIBSMun
(+) vCBS
(+) vIS
- Exceção 1: Em 2026 não somar vIBSUF, vIBSMun, vCBS, vIS,
vTotIBSMonoItem, vTotCBSMonoItem. * A confirmar.
- Nota: Implementação Futura.

```
Obrig. 1105 Rejeição: Valor total do Item (vItem) difere do somatório dos
valores que o compõem
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
Grupo VC. Referenciamento de item de outro Documento Fiscal Eletrônico - DF-e
Campo Modelo Regra de Validação Aplic. Msg Descrição Erro
VC02- 10 55/65 Item referenciado em duplicidade.
```
- Informado mais de uma vez uma mesma chave de acesso (tag:
    DFeReferenciado/chaveAcesso) e item (tag: DFeReferenciado/nItem).

```
Obrig. 1072 Rejeição: Item referenciado em duplicidade
```
```
VC02- 20 55/65 Um único documento fiscal deve ser referenciado:
```
- Informadas chaves de acesso diferentes em documento referenciado
    (tag: DFeReferenciado/chaveAcesso).

```
Obrig. 1072 Rejeição: Mais de um documento fiscal referenciado
```
```
Grupo W03. Total da NF-e - IBS / CBS / IS
Campo Modelo Regra de Validação Aplic. Msg Descrição Erro
W31- 10 55/65 O grupo de totais do IBS, CBS e IS (gIBSCBSTot) só deve ser informado
se existir pelo menos uma ocorrência de IBS, CBS ou Imposto Seletivo
nos itens.
```
```
Obrig. 1073 Rejeição: Total de IBS, CBS e IS só deve ser informado se
existir IBS/CBS declarado nos itens do DFe
```
```
W33- 10 55/65 O total da BC do imposto seletivo deverá ser a soma dos campos vBC
(tag: gIS/vBCIS) informados nos itens.
Nota: Implementação Futura.
```
```
Obrig. 1074 Rejeição: Total da BC do Imposto Seletivo difere da soma
dos itens
```
```
W34- 10 55/65 O total do Imposto Seletivo deverá ser a soma dos campos vIS (tag:
gIS/vIS) informados nos itens.
Nota: Implementação Futura.
```
```
Obrig. 1075 Rejeição: Total do Imposto Seletivo difere da soma dos itens
```
```
W35- 10 55/65 O total da BC do IBS e da CBS de deverá ser a soma dos campos vBC
(tag: gIBSCBS/vBC) informados nos itens
```
```
Obrig. 1076 Rejeição: Total da BC do IBS e da CBS difere da soma dos
itens
W38- 10 55/65 O total do diferimento do IBS UF deverá ser a soma do campo vDif do IBS
UF informados nos itens
```
```
Obrig. 1077 Rejeição: Total de Diferimento do IBS UF difere da soma dos
itens
W39- 10 55/65 O total devolvido do IBS UF deverá ser a soma do campo vDevTrib do IBS
UF informados nos itens
```
```
Obrig. 1078 Rejeição: Total Devolvido do IBS UF difere da soma dos
itens
W41- 10 55/65 O total do IBS UF deverá ser a soma do campo vIBSUF informados nos
itens
```
```
Obrig. 1080 Rejeição: Total de IBS UF difere da soma dos itens
```
```
W43- 10 55/65 O total do Diferimento do IBS Municipal deverá ser a soma do campo vDif
do IBS Municipal informados nos itens
```
```
Obrig. 1081 Rejeição: Total de Diferimento do IBS Municipal difere da
soma dos itens
W44- 10 55/65 O total devolvido do IBS Municipal deverá ser a soma do campo vDevTrib
do IBS Municipal informados nos itens
```
```
Obrig. 1082 Rejeição: Total Devolvido do IBS Municipal difere da soma
dos itens
W46- 10 55/65 O total do IBS Municipal deverá ser a soma do campo vIBSMun
informados nos itens
```
```
Obrig. 1084 Rejeição: Total de IBS Municipal difere da soma dos itens
```
```
W47- 10 55/65 O total do IBS deverá ser a soma do IBS das UF e do IBS Municipal
(vIBSUF + vIBSMun)
```
```
Obrig. 1085 Rejeição: Total do IBS difere da soma do IBS UF e IBS
Municipal
W48- 10 55/65 O total do Crédito Presumido do IBS deverá ser a soma do campo
vCredPres do IBS informados nos itens
```
```
Obrig. 1086 Rejeição: Total de Crédito Presumido do IBS UF difere da
soma dos itens
W51- 10 55/65 O total do Crédito Presumido da CBS deverá ser a soma do campo
vCredPres da CBS informados nos itens
```
```
Obrig. 1087 Rejeição: Total de Crédito Presumido da CBS difere da
soma dos itens
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
Campo Modelo Regra de Validação Aplic. Msg Descrição Erro
W53- 10 55/65 O total do Diferimento da CBS deverá ser a soma do campo vDif da CBS
informados nos itens
```
```
Obrig. 1088 Rejeição: Total de Diferimento da CBS difere da soma dos
itens
W54- 10 55/65 O total devolvido da CBS deverá ser a soma do campo vDevTrib da CBS
informados nos itens
```
```
Obrig. 1089 Rejeição: Total Devolvido da CBS difere da soma dos itens
```
```
W56- 10 55/65 O total da CBS deverá ser a soma do campo vCBS informados nos itens Obrig. 1091 Rejeição: Total de CBS difere da soma dos itens
W58- 10 55/65 O total do IBS monofásico deverá ser a soma dos campos vTOTIBSMono
informados nos itens
```
```
Obrig. 1092 Rejeição: Total do IBS monofásico difere da soma dos itens
```
```
W59- 10 55/65 O total da CBS monofásica deverá ser a soma dos campos
vTOTCBSMono informados nos itens
```
```
Obrig. 1093 Rejeição: Total da CBS monofásica difere da soma dos itens
```
```
W60- 10 55/65 O total geral do DFe (tag: vTotNF) deverá ser a soma do total dos itens
(vItem, tag VB 01 )
```
```
Obrig. 1094 Rejeição: Total da NFe difere da soma do total dos itens da
Nota
```
```
Grupo 3A. Banco de Dados: NF-e Referenciada
Campo Modelo Regra de Validação Aplic. Msg Descrição Erro
3BA02- 10 55 Para cada NF-e referenciada (tag:refNFe), se a UF da Chave de Acesso
referenciada for igual a UF do Emitente:
```
- Acessar BD NFE com Chave de Acesso referenciada (se mod=55)
- NF-e referenciada inexistente
Exceção: A NF-e referenciada pode não existir no caso de Emissão em
    Contingência (tpEmis = 2, 4 ou 5) (NT 2013/003), desde que a Chave
    de Acesso da NF-e referenciada tenha o Ano-Mês de Emissão inferior
    a 1 mês da data atual ou desde que exista o EPEC (NT 2021.004).
    Esta exceção não se aplica para NF-e com finalidade de crédito (tag:
    finNF-e = 5).
Observação 1: A exceção acima não se aplica para “finNFe=2" (NF-e
    Complementar).
Observação 2: Regra de validação se aplica obrigatoriamente para
    “finNFe=5" (NF-e de crédito).

```
Facult 267 Rejeição: Chave de Acesso referenciada inexistente [nRef:
xxx]
```
```
3BA02- 70 55 Se NF-e de crédito (tag: finNFe=5) e tpNFCredito=”04-Multa e juros”:
```
- Chave de acesso referenciada deve estar autorizada e ter finalidade
    de débito (tag: finNFe=6) e ter tpNFDebito=“04-Multa e juros”).
Nota Explicativa: Em operação entre empresas, o fornecedor pode
    cobrar multa e juros do cliente e no valor cobrado de juros e multas,
    incide o IBS e a CBS. O cliente ao pagar este valor adicional pode se
    creditar do imposto pago. Nesta situação pode ser emitida uma nota
    com o acerto dos juros e multas. A cobrança de juros e multas pode
    ser proveniente de atraso de pagamento no boleto bancário, sem que
    haja uma nota para juros e multas. Nessa situação deve haver um
    Evento para que o adquirente garanta o crédito adicional, gerando um
    débito para o outro contribuinte.

```
Obrig. 1095 Rejeição: Nota de crédito de multa/juros deve ter NF-e de
débito de multa/juros referenciada
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

8. Eventos

8.1. Lista de eventos

```
Esta NT cria os eventos a seguir para a apuração do IBS e da CBS.
CÓDIGO EVENTO Autor
112110 Informação de efetivo pagamento integral para liberar crédito presumido do adquirente Emitente
211110 Solicitação de Apropriação de crédito presumido Destinatário
211120 Destinação de item para consumo pessoal Emitente/Destinatário
211124 Perecimento, perda, roubo ou furto Destinatário
211128 Aceite de débito na apuração por emissão de nota de crédito Destinatário
211130 Imobilização de Item Destinatário
211140 Solicitação de Apropriação de Crédito de Combustível Destinatário
211150 Solicitação de Apropriação de Crédito para bens e serviços que dependem de atividade do adquirente Destinatário
212110 Manifestação sobre Pedido de Transferência de Crédito de IBS em Operações de Sucessão Sucessora
212120 Manifestação sobre Pedido de Transferência de Crédito CBS em Operações de Sucessão Sucessora
412120 Manifestação do Fisco sobre Pedido de Transferência de Crédito de IBS em Operações de Sucessão Fisco
412130 Manifestação do Fisco sobre Pedido de Transferência de Crédito de CBS em Operações de Sucessão Fisco
```
```
Criado também um código de Evento de Cancelamento genérico que será utilizado para o cancelamento de qualquer um dos Eventos citados
anteriormente nesta Nota Técnica. Neste Evento deve ser informado o código do Evento a ser cancelado.
CÓDIGO EVENTO Autor
110001 Cancelamento de Evento Idem ao Autor do Evento que está sendo cancelado
```
8.2. Registro de Eventos

O Web Service de Registro de Evento possui uma parte geral, complementada por uma área específica para cada tipo de evento.

```
A parte geral se mantém a mesma para envio de todos os eventos e está especificada na seção 5.8 do Manual de Orientação do Contribuinte e na
seção Web Service – NFeRecepcaoEvento – Parte Geral do MOC Online.
```
```
A mensagem de retorno será modificada, conforme apresentado na seção 8.3 e o leiaute das partes específicas dos novos eventos está
especificado nas seções seguintes.
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

8.3. Leiaute Mensagem de Retorno da Parte Geral

```
O Web Service de Registro de Evento possui uma interface genérica complementada por uma área específica para cada tipo de evento. Segue o
leiaute da mensagem de retorno (resposta).
```
```
Schema XML: retEnvEventoNFe_v1.0.xsd
# Campo Ele Pai Tipo Ocor. Tam. Descrição / Observações
R01 envEvento Raiz - - - - TAG raiz
R02 versão A R01 N 1 - 1 2v2 Versão do leiaute
R03 idLote E R01 N 1 - 1 1 - 15 Idem a mensagem de entrada
R04 tpAmb E R01 N 1 - 1 1 Idem a mensagem de entrada
R05 verAplic E R01 C 1 - 1 1 - 20 Versão da aplicação que processou o evento.
R06 cOrgao E R01 N 1 - 1 2 Órgão da recepção do Evento, idem a mensagem de entrada.
R07 cStat E R01 N 1 - 1 3 Código do status da resposta para o Lote de Eventos. Se não tiver erro, será
retornado: “128-Lote de Evento Processado“
R08 xMotivo E R01 C 1 - 1 1 - 255 Descrição do status da resposta
R09 retEvento G R01 - 0 - 20 - Grupo do resultado do processamento para cada Evento
R10 versão A R09 N 1 - 1 2v2 Versão do leiaute
R11 infEvento G R09 - 1 - 1 - Grupo de informações do registro do Evento.
R12 id ID R11 C 0 - 1 17 Identificador da TAG a ser assinada, somente deve ser informado se o órgão de
registro assinar a resposta. No caso de assinatura, preencher com o número do
protocolo, precedido pela literal “ID”.
R13 tpAmb E R11 C 1 - 1 1 Idem a mensagem de entrada
R14 verAplic E R11 N 1 - 1 1 - 20 Versão da aplicação que registrou o Evento, utilizar literal que permita a
identificação do órgão, como a sigla da UF ou do órgão.
R15 cOrgao E R11 N 1 - 1 2 Idem a mensagem de entrada
R16 cStat E R11 N 1 - 1 4 Código do status da resposta.
R17 xMotivo E R11 C 1 - 1 1 - 255 Descrição do status da resposta
R18 chNFe E R11 N 1 - 1 44 Idem a mensagem de entrada
R19 tpEvento E R11 N 0 - 1 6 Idem a mensagem de entrada
R20 xEvento E R11 C 0 - 1 5 - 60 Idem a mensagem de entrada
R21 nSeqEvento E R11 N 0 - 1 1 - 2 Idem a mensagem de entrada
R50 dhRegEvento E R11 D 1 - 1 - Data e hora de registro do evento no formato AAAA-MM-DDTHH:MM:SS TZD
(formato UTC). Se o evento for rejeitado informar a data e hora de recebimento do
evento.
R51 nProt E R11 N 0 - 1 15 Número Protocolo do Evento 1 posição (1- Secretaria da Fazenda Estadual, 2-RFB,
3 - SVRS), 2 posições para o código da UF, 2 posições para o ano e 10 posições
para o sequencial no ano.
P91 Signature G R09 XML 1 - 1 1 Assinatura Digital do documento XML, a assinatura deverá ser aplicada no
elemento infEvento.
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

8.4. Evento: Informação de efetivo pagamento integral para liberar crédito presumido do adquirente

```
Função: Permitir que o emitente da NFe informe o efetivo pagamento integral a fim de liberar crédito presumido do adquirente
Modelo: NF-e modelo 55
Autor do Evento: Emitente da NFe
Código do Tipo de Evento: 112110
```
8.4.1. Leiaute Mensagem de Entrada

```
Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos
especificada na seção 5.8 do MOC.
```
```
Schema XML: envEventoNFe_v9.99.xsd
Schema XML - parte específica: e112110_v1.00.xsd
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 - Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 85 Descrição do evento: "Informação de efetivo pagamento integral para liberar crédito presumido do adquirente"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código do Órgão Autor do Evento. Informar o Código da UF para este Evento.
P21 tpAutor E P17 N 1 - 1 1 Informar 1=Empresa emitente.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco; 6=RFB; 9=Outros Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 indQuitacao E P17 N 1 - 1 1 Indicador de efetiva quitação do pagamento integral referente a NFe referenciada. Valor deve ser igual a "1"
```
8.4.2. Leiaute Mensagem de Retorno

```
Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral, especificado
no item 7.2.1.
```
8.5. Evento: Solicitação de Apropriação de crédito presumido

```
Função: Evento a ser gerado pelo adquirente em relação às notas fiscais de aquisição de emissão de terceiros e que lhe gerem o direito à
apropriação de crédito presumido.
Autor: Adquirente/Destinatário (quando os dois estiverem preenchidos, devem ser iguais) da nota fiscal
Modelo: NF-e modelo 55
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

Código do Tipo de Evento: 211110

8.5.1. Leiaute Mensagem de Entrada

```
Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos
especificada na seção 5.8 do MOC.
```
```
Schema XML: envEventoNFe_v9.99.xsd
Schema XML - parte específica: e 211110 _v1.00.xsd
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 47 Descrição do evento: "Solicitação de Apropriação de crédito presumido"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código do Órgão Autor do Evento. Informar o Código da UF para este Evento.
P21 tpAutor E P17 N 1 - 1 1 Informar 1=Empresa emitente.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco;
6=RFB; 9=Outros Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 gCredPres G P17 1 - 990 Informações de crédito presumido por item
P24 nitem A P23 N 1 - 1 1 - 3 Corresponde ao atributo “nItem” do elemento “det” do documento referenciado.
P25 vBC A P23 N 1 - 1 1 Valor do base de cálculo do item
P26 gIBS G P23 0 - 1 Grupo de Informações do Crédito Presumido do IBS
P27 cCredPres E P27 N 1 - 1 2 Código de Classificação do Crédito presumido, conforme tabela de
CÓDIGO DE CLASSIFICAÇÃO DO CRÉDITO PRESUMIDO
P28 pCredPres E P27 N 1 - 1 3v2- 4 Percentual do Crédito Presumido
P29 vCredPres E P27 N 1 - 1 13v2 Valor do Crédito Presumido
```
```
8.5.2. Leiaute Mensagem de Retorno
Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral, especificado
no item 7.2.1.
```
8.5.3. Validação das Regras de Negócio **–** Específicas

Serão aplicadas as regras de validação gerais apresentadas no item 5.8.4 do MOC e as regras de negócio específicas listadas a seguir.

```
# Regra de Validação Aplic. Msg Descrição Erro
Banco de Dados: NF-e
2P21- 10 Se tpAutor=2-Empresa Destinatário:
```
- CNPJ/CPF do Autor diverge do CNPJ/CPF do Destinatário

```
Obrig. 575 Rejeição: Autor do evento diverge do destinatário da NF-e
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
# Regra de Validação Aplic. Msg Descrição Erro
da NF-e
2P24- 10 Acessar BD e verificar se número do item do evento (tag:
gCredPres/nItem) informado existe na NFe referenciada (tag:
det/nItem)
```
```
Obrig. 1096 Rejeição: Número de item não existe na NFe
```
8.6. Evento: Destinação de item para consumo pessoal

```
Função: Permitir ao adquirente informar quando uma aquisição for destinada para o consumo de pessoa física, hipótese em que não haverá direito
à apropriação de crédito. Evento a ser registrado após a emissão da nota de bens destinados para uso e consumo pessoal.
Uma mesma NFe de aquisição pode receber vários Eventos desse tipo, com nSeqEvento diferentes (eventos cumulativos).
Modelo: NF-e modelo 55
Autor do Evento: Destinatário da NF-e
Código do Tipo de Evento: 211120
```
8.6.1. Leiaute Mensagem de Entrada

```
Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos
especificada na seção 5.8 do MOC.
```
```
Schema XML: envEventoNFe_v9.99.xsd
Schema XML - parte específica: e 211120 _v1.00.xsd
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 39 Descrição do evento: "Destinação de item para consumo pessoal"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código do Órgão Autor do Evento. Informar o Código da UF para este Evento.
P21 tpAutor E P17 N 1 - 1 1 Caso NF-e de Importação, informar 1=Empresa Emitente.
Demais casos, informar 2=Empresa destinatária.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 gConsumo G P17 1 - 990 Informações por item da NF-e de Aquisição
Nota: a quantidade de ocorrências não pode ser maior que a quantidade de itens da NF-e de aquisição.
P24 nItem A P23 N 1 - 1 1 - 3 Corresponde ao atributo “nItem” do elemento “det” da NF-e de aquisição.
P25 vIBS A P23 N 1 - 1 13v2 Valor do IBS na nota de aquisição correspondente à quantidade destinada a uso e consumo pessoal
P26 vCBS A P23 N 1 - 1 13v2 Valor da CBS na nota de aquisição correspondente à quantidade destinada a uso e consumo pessoal
P27 gControleEstoque G P23 1 - 1 Informações de quantidade de estoque influenciadas pelo evento
P28 qConsumo E P2 7 N 1 - 1 11v0- 4 Informar a quantidade para consumo de pessoa física
P29 uConsumo E P2 7 C 1 - 1 1 - 6 Informar a unidade relativa ao campo gConsumo
P30 Sequência XML G P23 1 - 1 Informações por item da NF-e de Uso e Consumo Pessoal
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P3 1 refNF E P30 C 1 - 1 44 Informar a chave da nota (NFe ou NFCe) emitida para o fornecimento nos casos em que a legislação
obriga a emissão de documento fiscal.
P3 2 nItemRefNFe E P30 N 1 - 1 1 - 3 Corresponde ao “nItem” da refNFe
```
8.6.2. Leiaute Mensagem de Retorno

```
Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral, especificado
no item 7.2.1.
```
8.6.3. Validação das Regras de Negócio **–** Específicas

Serão aplicadas as regras de validação gerais apresentadas no item 5.8.4 do MOC e as regras de negócio específicas listadas a seguir.

```
# Regra de Validação Aplic. Msg Descrição Erro
2P12- Acesso ao BD NFe (chave=Chave de Acesso da NFe de Aquisição):
```
- NF-e de aquisição inexistente.
Exceção 01 : NF-e de Aquisição pode não existir se UF da Chave
    Acesso diverge da Sefaz Autorizadora.
Exceção 2: A NF-e referenciada pode não existir no caso de
    Emissão em Contingência (tpEmis = 2, 4 ou 5), desde que a
    Chave de Acesso da NF-e referenciada tenha o Ano-Mês de
    Emissão inferior a 1 mês da data atual ou desde que exista o
    EPEC.

```
Obrig. 1110 Rejeição: NF-e de Aquisição inexistente
```
```
2P12- 20 - Se NF-e de aquisição é de importação:
```
- tpAutor do Evento deve ser “1=Empresa Emitente”

```
Obrig. 466 Rejeição: Evento com Tipo de Autor incompatível:
```
```
2P12- 30 - Se NF-e de aquisição não é de importação:
```
- tpAutor do Evento deve ser “2=Empresa Destinatário”

```
Obrig. 466 Rejeição: Evento com Tipo de Autor incompatível:
```
```
2P21- 10 - Se tpAutor=2-Empresa Destinatário:
```
- CNPJ/CPF do Autor do Evento diverge do CNPJ/CPF do
    Destinatário da NF-e de Aquisição

```
Obrig. 575 Rejeição: Autor do evento diverge do destinatário da NF-e
```
```
2P21- 30 - Se tpAutor=1-Empresa Emitente:
```
- CNPJ/CPF do Autor do Evento diverge do CNPJ/CPF do
    Emitente da NF-e de Aquisição

```
Obrig. 574 Rejeição: O autor do evento diverge do emissor da NF-e
```
```
2P24- 10 - Número do item do evento (tag: gconsumo/nItem) informado não
existe na NFe de Aquisição (tag: det/nItem)
```
```
Obrig. 1096 Rejeição: Número de item não existe na NFe
```
```
2P25- 10 - Valor do IBS do evento (tag: gConsumo/vIBS) maior que o valor
do IBS do item informado na NFe de Aquisição
```
```
Obrig. 1097 Rejeição: O valor do IBS do item não pode ser maior que o valor do IBS do
respectivo item na NFe.
2P26- 10 - Valor da CBS do evento (tag: gConsumo/vCBS) maior que o
valor da CBS do item informado na NFe de Aquisição
```
```
Obrig. 1098 Rejeição: O valor da CBS do item não pode ser maior que o valor da CBS do
respectivo item na NFe.
2P28- 10 - Se unidade de consumo do Evento for igual a unidade do item Obrig. 1099 Rejeição: A quantidade de consumo não pode ser maior que a quantidade do
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
# Regra de Validação Aplic. Msg Descrição Erro
da nota de Aquisição (tag: det/prod/uCom):
```
- Quantidade de consumo do item informado (tag:
    gControleEstoque/qConsumo) maior que a quantidade no
    item da NFe (tag: det/prod/qCom)

```
respectivo item na NFe (qCom)
```
```
2P3 1 - 10 Acesso ao BD NFe ou NFCe (chave=refNFe):
```
- Chave de acesso inexistente
Exceção 01: NF-e pode não existir se UF da Chave Acesso diverge
    da Sefaz Autorizadora.
Exceção 2: A NF-e referenciada pode não existir no caso de
Emissão em Contingência (tpEmis = 2, 4 ou 5), desde que a Chave
de Acesso tenha o Ano-Mês de Emissão inferior a 1 mês da data
atual ou desde que exista o EPEC.

```
Obrig. 494 Rejeição: Chave de Acesso inexistente
[chNFe:99999999999999999999999999999999999999999999]
```
```
2P3 1 - 20 Se refNFe existe e modelo = 55:
```
- refNFe não é de Saída

```
Obrig. 1111 Rejeição: Nota Fiscal referenciada não é de Saída
```
```
2P3 2 - 10 Se refNFe existe:
```
- nItemRefNFe da refNFe com cClassTrib diferente de “xxxx”

```
Obrig. 1112 Rejeição: cClassTrib da Nota Fiscal referenciada inválido
```
8.7. Evento: Perecimento, perda, roubo ou furto

```
Função: Permitir ao adquirente informar quando uma aquisição for objeto de roubo, perda, furto ou perecimento, hipótese em que não haverá direito
à apropriação de crédito.
Modelo: NF-e modelo 55
Autor do Evento: Destinatário da NF-e
Código do Tipo de Evento: 211124
```
8.7.1. Leiaute Mensagem de Entrada

```
Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos
especificada na seção 5.8 do MOC.
```
```
Schema XML: envEventoNFe_v9.99.xsd
Schema XML - parte específica: e 211124 _v1.00.xsd
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 39 Descrição do evento: "Destinação de item para consumo pessoal"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código do Órgão Autor do Evento. Informar o Código da UF para este Evento.
P21 tpAutor E P17 N 1 - 1 1 Informar 2=Empresa destinatária.
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco; 6=RFB; 9=Outros Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 gPerecimento G P17 1 - 990 Informações por item da Nota de Aquisição
P24 nItem A P23 N 1 - 1 1 - 3 Corresponde ao atributo “nItem” do elemento “det” do documento referenciado.
P25 vIBS A P23 N 1 - 1 13v2 Valor do IBS na nota de aquisição correspondente à quantidade que foi objeto de roubo, perda, furto ou perecimento
P26 vCBS A P23 N 1 - 1 13v2 Valor da CBS na nota de aquisição correspondente à quantidade que foi objeto de roubo, perda, furto ou perecimento
P27 gControleEstoque G P23 1 - 1 Informações de quantidade de estoque influenciadas pelo evento
P28 qPerecimento E P27 N 1 - 1 11v0- 4 Informar a quantidade que foi objeto de roubo, perda, furto ou perecimento
P29 uPerecimento E P27 C 1 - 1 1 - 6 Informar a unidade relativa ao campo qPerecimento
```
8.7.2. Leiaute Mensagem de Retorno

```
Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral, especificado
no item 7.2.1.
```
```
8.7.3. Validação das Regras de Negócio – Específicas
Serão aplicadas as regras de validação gerais apresentadas no item 5.8.4 do MOC e as regras de negócio específicas listadas a seguir.
```
```
# Regra de Validação Aplic. Msg Descrição Erro
```
8.8. Evento: Aceite de débito na apuração por emissão de nota de crédito

```
Função: Permitir ao destinatário informar que concorda com os valores constantes em nota de crédito emitida pelo fornecedor ou pelo adquirente
que serão lançados a débito na apuração assistida de IBS e CBS
Modelo: NF-e modelo 55
Autor do Evento: Destinatário da NF-e
Código do Tipo de Evento: 211128
```
8.8.1. Leiaute Mensagem de Entrada

```
Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos
especificada na seção 5.8 do MOC.
```
```
Schema XML: envEventoNFe_v9.99.xsd
Schema XML - parte específica: e 211128 _v1.00.xsd
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 85 Descrição do evento: "Manifestação sobre Pedido de Transferência de Crédito de IBS em Operações de Sucessão"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código da UF do emitente do Evento
P21 tpAutor E P17 N 1 - 1 1 Informar 2=Empresa destinatária.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 indAceitacao A P17 N 1 - 1 1 Indicador de concordância com o valor da nota de crédito que lançaram IBS e CBS na apuração assistida. Valores: 0 = não
aceite; 1 = aceite.
```
8.8.2. Leiaute Mensagem de Retorno

```
Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral, especificado
no item 7.2.1.
```
8.8.3. Validação das Regras de Negócio **–** Específicas

Serão aplicadas as regras de validação gerais apresentadas no item 5.8.4 do MOC e as regras de negócio específicas listadas a seguir.

```
# Regra de Validação Aplic. Msg Descrição Erro
```
8.9. Evento: Imobilização de Item

```
Função: Evento a ser gerado pelo adquirente de bem, quando este for integrado ao seu ativo imobilizado, a fim de viabilizar a adequada
identificação, pelos sistemas da administração tributária, de prazo-limite para apreciação de eventuais pedidos de ressarcimento do respectivo
crédito, nos termos do art. 40, I da LC 214/2025.
Modelo: NF-e modelo 55
Autor do Evento: Destinatário da NF-e (Adquirente)
Código do Tipo de Evento: 211130
```
```
8.9.1. Leiaute Mensagem de Entrada
Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos
especificada na seção 5.8 do MOC.
```
```
Schema XML: envEventoNFe_v9.99.xsd
Schema XML - parte específica: e211130_v1.00.xsd
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 5 - 60 Descrição do evento: "Imobilização de Item"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código da UF do emitente do Evento
P21 tpAutor E P17 N 1 - 1 1 Informar 2=Empresa destinatária.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 gImobilizacao G P17 1 - 990 Informações de itens integrados ao ativo imobilizado
P24 nitem A P23 N 1 - 1 1 - 3 Corresponde ao atributo “nItem” do elemento “det” do documento referenciado.
P25 vIBS A P23 N 1 - 1 13v2 Valor do IBS relativo à imobilização
P26 vCBS A P23 N 1 - 1 13v2 Valor da CBS relativo à imobilização
P26 gControleEstoque G P26 1 - 1 Informações de crédito presumido por item
P27 qImobilizado E P26 N 1 - 1 11v0- 4 Informar a quantidade do item a ser imobilizado
P28 uImobilizado E P26 C 1 - 1 1 - 6 Informar a unidade relativa ao campo qImobilizado
```
8.9.2. Leiaute Mensagem de Retorno

```
Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral, especificado
no item 7.2.1.
```
8.9.3. Validação das Regras de Negócio **–** Específicas

Serão aplicadas as regras de validação gerais apresentadas no item 5.8.4 do MOC e as regras de negócio específicas listadas a seguir.

```
# Regra de Validação Aplic. Msg Descrição Erro
Banco de Dados: NF-e
2P21- 10 Se tpAutor=2-Empresa Destinatário:
```
- CNPJ/CPF do Autor diverge do CNPJ/CPF do
Destinatário da NF-e

```
Obrig. 575 Rejeição: Autor do evento diverge do destinatário da NF-e
```
```
2P24- 10 Acessar BD e verificar se número do item do evento (tag:
gImobilizacao/nItem) informado existe na NFe referenciada
(tag: det/nItem)
```
```
Obrig. 1096 Rejeição: Número de item não existe na NFe
```
```
2P25- 10 Acessar BD e verificar se valor do IBS do evento (tag:
gImobilizacao/vIBS) é maior que o valor do IBS do item
informado na NFe
```
```
Obrig. 1097 Rejeição: O valor do IBS do item não pode ser maior que o valor do IBS do
respectivo item na NFe.
```
```
2P26- 10 Acessar BD e verificar se valor da CBS do evento (tag:
gImobilizacao /vCBS) é maior que o valor da CBS do item
informado na NFe
```
```
Obrig. 1098 Rejeição: O valor da CBS do item não pode ser maior que o valor da CBS
do respectivo item na NFe.
```
```
2P28- 10 Se unidade for igual a unidade da nota, acessar BD e verificar
se a quantidade de consumo do item informado (tag:
gControleEstoque/qImobilizada) é maior que a quantidade no
```
```
Obrig. 1100 Rejeição: A quantidade de consumo não pode ser maior que a quantidade
do respectivo item na NFe (qCom)
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
# Regra de Validação Aplic. Msg Descrição Erro
item da NFe (tag: det/prod/qCom)
```
8.10. Evento: Solicitação de Apropriação de Crédito de Combustível

```
Função: Evento a ser gerado pelo adquirente de combustível listado no art. 172 da LC 214/2025 e que pertença à cadeia produtiva desses
combustíveis, para solicitar a apropriação de crédito referente à parcela que for consumida em sua atividade comercial.
Modelo: NF-e modelo 55
Autor do Evento: Destinatário da NF-e (Adquirente de combustível que faça parte da cadeia produtiva de combustíveis)
Código do Tipo de Evento: 211140
```
8.10.1. Leiaute Mensagem de Entrada

```
Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos
especificada na seção 5.8 do MOC.
```
```
Schema XML: envEventoNFe_v9.99.xsd
Schema XML - parte específica: e211140_v1.00.xsd
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 - Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 52 Descrição do evento: "Solicitação de Apropriação de Crédito de Combustível"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código da UF do emitente do Evento
P21 tpAutor E P17 N 1 - 1 1 Informar 2=Empresa destinatária.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco; 6=RFB; 9=Outros Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 gConsumoComb G P17 1 - 990 Informações de consumo de combustíveis
P24 nitem A P23 N 1 - 1 1 - 3 Corresponde ao atributo “nItem” do elemento “det” do documento referenciado.
P25 vIBS A P23 N 1 - 1 13v2 Valor do IBS relativo ao consumo de combustível na nota de aquisição
P26 vCBS A P23 N 1 - 1 13v2 Valor da CBS relativo ao consumo de combustível na nota de aquisição
P27 gControleEstoque G P26 1 - 1 Informações de quantidade por item
P28 qComb A P27 N 1 - 1 11v0- 4 Informar a quantidade de consumo do item
P28 uComb G P27 N 1 - 1 1 Informar a unidade relativa ao campo qComb
```
8.10.2. Leiaute Mensagem de Retorno

```
Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral, especificado
no item 7.2.1.
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

8.10.3. Validação das Regras de Negócio **–** Específicas

Serão aplicadas as regras de validação gerais apresentadas no item 5.8.4 do MOC e as regras de negócio específicas listadas a seguir.

```
# Regra de Validação Aplic. Msg Descrição Erro
Banco de Dados: NF-e
2P21- 10 Se tpAutor=2-Empresa Destinatário:
```
- CNPJ/CPF do Autor diverge do CNPJ/CPF do Destinatário da NF-e

```
Obrig. 575 Rejeição: Autor do evento diverge do destinatário da NF-e
```
```
2P24- 10 Acessar BD e verificar se número do item do evento (tag:
gConsumoComb/nItem) informado existe na NFe referenciada (tag:
det/nItem)
```
```
Obrig. 1096 Rejeição: Número de item não existe na NFe
```
```
2P25- 10 Acessar BD e verificar se valor do IBS do evento (tag:
gConsumoComb/vIBS) é maior que o valor do IBS do item informado na
NFe
```
```
Obrig. 1097 Rejeição: O valor do IBS do item não pode ser maior que o valor do
IBS do respectivo item na NFe.
```
```
2P26- 10 Acessar BD e verificar se valor da CBSdo evento (tag:
gConsumoComb/vCBS) é maior que o valor da CBS do item informado na
NFe
```
```
Obrig. 1098 Rejeição: O valor da CBS do item não pode ser maior que o valor da
CBS do respectivo item na NFe.
```
```
2P28- 10 Se unidade de consumo for igual a unidade da nota, acessar BD e verificar
se a quantidade de consumo do item informado (tag:
gConsumoComb/qComb) é maior que a quantidade no item da NFe (tag:
det/prod/qCom)
```
```
Obrig. 1101 Rejeição: A quantidade do item a ser imobilizado não pode ser maior
que a quantidade do respectivo item na NFe (qCom)
```
8.11. Evento: Solicitação de Apropriação de Crédito para bens e serviços que dependem de atividade do

adquirente

```
Função: Evento a ser gerado pelo adquirente para apropriação de crédito de bens e serviços que dependam da sua atividade
Modelo: NF-e modelo 55
Autor do Evento: Destinatário da NFe (adquirente).
Código do Tipo de Evento: 211150
```
8.11.1. Leiaute Mensagem de Entrada

```
Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos
especificada na seção 5.8 do MOC.
```
```
Schema XML: envEventoNFe_v9.99.xsd
Schema XML - parte específica: e211150_v1.00.xsd
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 Detalhes do Evento
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 98 Descrição do evento: "Solicitação de Apropriação de Crédito para bens e serviços que dependem de atividade do adquirente"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código da UF do emitente do Evento
P21 tpAutor E P17 N 1 - 1 1 Informar 2=Empresa destinatária.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco; 6=RFB; 8= Empresa sucessora; 9=Outros
Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 gCredito G P17 1 - 990 Informações de crédito
P24 nitem A P23 N 1 - 1 1 - 3 Corresponde ao atributo “nItem” do elemento “det” do documento referenciado.
P25 vCredIBS E P23 N 1 - 1 13v2 Valor da solicitação de crédito a ser apropriado de IBS
P26 vCredCBS E P23 N 1 - 1 13v2 Valor da solicitação de crédito a ser apropriado de CBS
```
8.11.2. Leiaute Mensagem de Retorno

```
Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral, especificado
no item 7.2.1.
```
8.11.3. Validação das Regras de Negócio **–** Específicas

Serão aplicadas as regras de validação gerais apresentadas no item 5.8.4 do MOC e as regras de negócio específicas listadas a seguir.

```
# Regra de Validação Aplic. Msg Descrição Erro
Banco de Dados: NF-e
2P21- 10 Se tpAutor=2-Empresa Destinatário:
```
- CNPJ/CPF do Autor diverge do CNPJ/CPF do Destinatário da NF-e

```
Obrig. 575 Rejeição: Autor do evento diverge do destinatário da NF-e
```
```
P24- 10 Acessar BD e verificar se número do item do evento (tag: gCredito/nItem)
informado existe na NFe referenciada (tag: det/nItem)
```
```
Obrig. 1096 Rejeição: Número de item não existe na NFe
```
```
P25- 10 Acessar BD e verificar se valor do crédito de IBS do evento (tag:
gCredito/vCredIBS) é maior que o valor do IBS do item informado na NFe
```
```
Obrig. 1097 Rejeição: O valor do crédito de IBS do item não pode ser
maior que o valor do IBS do respectivo item na NFe.
P26- 10 Acessar BD e verificar se valor do crédito de CBS do evento (tag:
gCredito/vCredCBS) é maior que o valor do IBS do item informado na NFe
```
```
Obrig. 1098 Rejeição: O valor do crédito de CBS do item não pode ser
maior que o valor da CBS do respectivo item na NFe.
```
8.12. Evento: Manifestação sobre Pedido de Transferência de Crédito de IBS em Operações de Sucessão

```
Função: Evento a ser gerado pela sucessora em relação às notas fiscais de transferência de crédito de outra sucessora da mesma empresa
sucedida para informar aceite da transferência de crédito de IBS.
Autor: Empresa sucessora
Modelo: NF-e modelo 55
Código do Tipo de Evento: 212110
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

8.12.1. Leiaute Mensagem de Entrada

```
Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos
especificada na seção 5.8 do MOC.
```
```
Schema XML: envEventoNFe_v9.99.xsd
Schema XML - parte específica: e212110.00.xsd
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 85 Descrição do evento: "Manifestação sobre Pedido de Transferência de Crédito de IBS em Operações de Sucessão"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código da UF do emitente do Evento
P21 tpAutor E P17 N 1 - 1 1 Informar 8=Empresa sucessora.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco; 6=RFB; 8= Empresa sucessora; 9=Outros Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 indAceitacao A P17 N 1 - 1 1 Indicador de aceitação do valor de transferência para a empresa que emitiu a nota referenciada.
Valores: 0=Não Aceite; 1=Aceite.
```
8.12.2. Leiaute Mensagem de Retorno

```
Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral, especificado
no item 7.2.1.
```
8.13. Evento: Manifestação sobre Pedido de Transferência de Crédito CBS em Operações de Sucessão

```
Função: Evento a ser gerado pela sucessora em relação às notas fiscais de transferência de crédito de outra sucessora da mesma empresa
sucedida para informar aceite da transferência de crédito de CBS.
Autor: Empresa sucessora
Modelo: NF-e modelo 55
Código do Tipo de Evento: 212120
```
8.13.1. Leiaute Mensagem de Entrada

```
Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos
especificada na seção 5.8 do MOC.
```
```
Schema XML: envEventoNFe_v9.99.xsd
Schema XML - parte específica: e 212120 _v1.00.xsd
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 82 Descrição do evento: "Manifestação sobre Pedido de Transferência de Crédito CBS em Operações de Sucessão"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código da UF do emitente do Evento
P21 tpAutor E P17 N 1 - 1 1 Informar 8=Empresa sucessora.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 indAceitacao A P17 N 1 - 1 1 Indicador de aceitação do valor de transferência para a empresa que emitiu a nota referenciada.
Valores: 0=Não Aceite; 1=Aceite.
```
8.13.2. Leiaute Mensagem de Retorno

```
Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral, especificado
no item 7.2.1.
```
8.14. Evento: Manifestação do Fisco sobre Pedido de Transferência de Crédito de IBS em Operações de

Sucessão

```
Função: Evento a ser gerado pelo fisco em relação às notas fiscais de transferência de crédito para informar aceite ou não aceite da transferência
de crédito de IBS.
Autor: Fisco
Modelo: NF-e modelo 55
Código do Tipo de Evento: 412120
```
8.14.1. Leiaute Mensagem de Entrada

```
Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos
especificada na seção 5.8 do MOC.
```
```
Schema XML: envEventoNFe_v9.99.xsd
Schema XML - parte específica: e 412120 _v1.00.xsd
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 94 Descrição do evento: "Manifestação do Fisco sobre Pedido de Transferência de Crédito de IBS em Operações de Sucessão"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código da UF do emitente do Evento
P21 tpAutor E P17 N 1 - 1 1 Informar 5=Fisco
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P23 indDeferimento A P17 N 1 - 1 1 Indicador de aceitação do valor de transferência para a empresa que emitiu a nota referenciada.
Valores: 0=Não Aceite; 1=Aceite.
P24 cMotivo E P17 N 1 - 1 1 1 – Falta de manifestação de todas as sucessoras; 2 – Outros.
P25 xMotivo E P17 C 1 - 1 500
```
8.14.2. Leiaute Mensagem de Retorno

```
Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral, especificado
no item 7.2.1.
```
8.15. Evento: Manifestação do Fisco sobre Pedido de Transferência de Crédito de CBS em Operações de

Sucessão

```
Função: Evento a ser gerado pelo fisco em relação às notas fiscais de transferência de crédito para informar aceite ou não aceite da transferência
de crédito de CBS.
Autor: Fisco
Modelo: NF-e modelo 55
Código do Tipo de Evento: 412130
```
8.15.1. Leiaute Mensagem de Entrada

```
Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos
especificada na seção 5.8 do MOC.
```
```
Schema XML: envEventoNFe_v9.99.xsd
Schema XML - parte específica: e 412130 _v1.00.xsd
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 94
P20 cOrgaoAutor E P17 N 1 - 1 2 Descrição do evento: "Manifestação do Fisco sobre Pedido de Transferência de Crédito de CBS em Operações de Sucessão"
P21 tpAutor E P17 N 1 - 1 1 Informar 5=Fisco.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 indDeferimento A P17 N 1 - 1 1 Indicador de aceitação do valor de transferência para a empresa que emitiu a nota referenciada.
Valores: 0=Não Aceite; 1 = Aceite.
P24 cMotivo E P17 N 1 - 1 1 1 – Falta de manifestação de todas as sucessoras; 2 – Outros.
P25 xMotivo E P17 C 1 - 1 500
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

8.15.2. Leiaute Mensagem de Retorno

```
Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral, especificado
no item 7.2.1.
```
8.16. Evento: Cancelamento de Evento

```
Função: Permitir que o autor de um Evento já autorizado possa proceder o seu cancelamento.
Modelo: NF-e modelo 55
Autor do Evento: O mesmo Autor do Evento que está sendo cancelado.
Tipo de Evento (Código - Descrição): 110001 - Cancelamento de Evento
```
8.16.1. Leiaute Mensagem de Entrada

```
Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos
especificada na seção 5.8 do MOC.
```
```
Schema XML: envEventoNFe_v9.99.xsd
Schema XML - parte específica: e110001_v1.00.xsd
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 - 1 - 1 Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 22 Informar “Evento de Cancelamento”
P20 cOrgaoAutor E P17 N 1 - 1 2 Código da UF do autor do Evento
P22 verAplic E P17 C 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 tpEventoAut E P17 N 1 - 1 6 Código do evento autorizado a ser cancelado
P24 nProtEvento E P17 N 1 - 1 15 Informar o número do Protocolo de Autorização do Evento a ser cancelado
```
```
8.16.2. Leiaute Mensagem de Retorno
Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral, especificado
no item 7.2.1.
```
8.16.3. Validação das Regras de Negócio **–** Específicas

Serão aplicadas as regras de validação gerais apresentadas no item 5.8.4 do MOC e as regras de negócio específicas listadas a seguir.

```
# Regra de Validação Aplic. Msg Descrição Erro
*** Banco de Dados: Evento 2
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

```
# Regra de Validação Aplic. Msg Descrição Erro
1P 06 - 10 Acesso BD de Eventos (Chave: Chave de Acesso, tpEventoAut,
nSeqEvento):
```
- Evento inexistente

```
Obrig. 459 Rejeição: Cancelamento de Evento inexistente
```
```
1P10- 10 - CNPJ/CPF do Autor do Evento de Cancelamento diverge do
CNPJ/CPF do Autor do Evento a ser cancelado
```
```
Obrig. 1113 Rejeição: Autor do Evento de Cancelamento diverge do Autor
do Evento a ser cancelado
1P2 4 - 10 - Número do Protocolo diverge Obrig. 460 Rejeição: Protocolo do Evento difere do cadastrado
```
9. DANFE

```
Alterações no DANFE para exibir informações relativas aos novos tributos estão em estudo, e serão publicadas em uma nova versão desta Nota
Técnica.
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2025.002-RTC

ANEXO I - NCM DO IMPOSTO SELETIVO

```
Link para a tabela:
https://docs.google.com/spreadsheets/d/1TnXQPmAgAyvOgSIznmw1oxdVmvUIqkYTMH0Xxbepby8/edit?usp=sharing
```
ANEXO II - CÓDIGO DE CLASSIFICAÇÃO TRIBUTÁRIA DO IMPOSTO SELETIVO

Tabela a ser publicada.

ANEXO III - CÓDIGO DE CLASSIFICAÇÃO TRIBUTÁRIA DO IBS E DA CBS

Tabela a ser publicada.

ANEXO IV - CÓDIGO DE CLASSIFICAÇÃO DO CRÉDITO PRESUMIDO

Tabela a ser publicada.
