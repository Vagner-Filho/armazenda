**Projeto Reforma Tributária do**

**Consumo – Adequações NF-e / NFC-e**

**RT Nota Técnica 2024.002 - Versão 1. 10**

**06 de dezembro de 2024**


## Sumário

- Reforma Tributária - Emenda Constitucional 132/
- NT 2024.002 Versão 1.
- Introdução 1. Sumário
- 1 Tipos Básicos da Tributação
- 2 Código Situação Tributária e Código de Classificação da Tributação
- 3 Finalidades débito e crédito da NF-e........................................................................................
- 4 Alterações no arquivo XML da NF-e
   - 4.1 Esquema gráfico do leiaute do IBS da UF, contemplando IBS, CBS e Imposto Seletivo.
   - 4.2 Inserção dos novos campos de informação no leiaute da NF-e e NFC-e
   - 4.3 Disposição dos campos do Imposto Seletivo
   - 4.4 Imposto de Bens e Serviços - IBS das UFs e Municípios
   - 4.5 Disposição dos campos do IBS da UF
   - 4.6 Disposição dos campos do IBS dos Municípios
   - 4.7 Disposição dos campos da CBS
   - 4.8 Disposição dos campos do IBS Monofásico
   - 4.9 Disposição dos campos totalizadores
   - 4.10 Indicadores de multa, juros e compra governamental e tipo de nota de crédito.
   - 4.11 Referenciamento de Item da NF-e
- 5 Protocolo da NFe
- 6 Leiaute da NF-e (Modelo 55 e 65)
   - Grupo B. Identificação da Nota Fiscal eletrônica
   - Grupo N01. ICMS Normal e ST
   - Grupo UB. Informações dos tributos IBS / CBS e Imposto Seletivo
   - Grupo VB. Referenciamento de item de outro Documento Fiscal Eletrônico - DF-e
   - Grupo W03. Total da NF-e - IBS / CBS / IS
- 7 Regras de Validação
   - Grupo B. Identificação da Nota Fiscal eletrônica
   - Grupo BA. Documento Fiscal Referenciado
   - Grupo UB. Informações dos tributos IBS / CBS e Imposto Seletivo
   - Grupo VB. Referenciamento de item de outro Documento Fiscal Eletrônico - DF-e
   - Grupo W03. Total da NF-e - IBS / CBS / IS
- 3A. Banco de Dados: NF-e Referenciada
- 8 Eventos
   - 8.1 Lista de eventos
   - 7.2 Registro de Eventos
      - 7.2.1 Leiaute Mensagem de Retorno da Parte Geral


## Reforma Tributária - Emenda Constitucional 132/

NT 2024.002 Versão 1. 10

```
7.2.2 Evento: Informação de efetivo pagamento integral para liberar crédito presumido do
adquirente .............................................................................................................................. 42
```
7.2.3 Evento: Solicitação de Apropriação de crédito presumido ............................................. 44

7.2.4 Evento: Destinação de item para consumo pessoal ...................................................... 46

7.2.5 Evento: Imobilização de Item ........................................................................................ 48

7.2.6 Evento: Solicitação de Apropriação de Crédito de Combustível .................................... 50

7.2.7 Evento: Solicitação de Apropriação de Crédito de fornecimento de bem móvel usado. 52

```
7.2.8 Evento: Solicitação de Apropriação de Crédito para bens e serviços que dependem de
atividade do adquirente .......................................................................................................... 54
```
```
7.2.9 Evento: Manifestação sobre Pedido de Transferência de Crédito de IBS em Operações
de Sucessão .......................................................................................................................... 55
```
```
7.2.10 Evento: Manifestação sobre Pedido de Transferência de Crédito CBS em Operações de
Sucessão ............................................................................................................................... 56
```
```
7.2.11 Evento: Manifestação do Fisco sobre Pedido de Transferência de Crédito de IBS em
Operações de Sucessão ........................................................................................................ 57
```
```
7.2.12 Evento: Manifestação do Fisco sobre Pedido de Transferência de Crédito de CBS em
Operações de Sucessão ........................................................................................................ 58
```
7.2.13 Evento: Evento de cancelamento de evento do emitente ............................................ 59

7.2.14 Evento: Evento de cancelamento de evento do destinatário ....................................... 61

7.2.15 Evento: Evento de cancelamento de evento da sucessora ......................................... 62

9 DANFE .................................................................................................................................. 64


## Reforma Tributária - Emenda Constitucional 132/

NT 2024.002 Versão 1. 10

Controle de Versões

## Versão Publicação Descrição

```
1.00 07 /202 4 Criação deste manual como documento com campos e regras para informação do Imposto
de Bens e Serviços - IBS, Contribuição de Bens e Serviços - CBS e Imposto Seletivo - IS.
```
## 1. 10 12 /2024 Inserção de campos de controle e criação de eventos para utilização na apuração do IBS e

## da CBS

Histórico de Alterações / Cronograma

## Versão Histórico de atualizações Implantação

## Teste

## Implantação

## Produção

## 1.

## Versão inicial com a previsão dos campos do imposto seletivo,

## IBS e CBS na NF-e e NFC-e.

## 01/09/2025 31/10//

## 1. 10

## Versão contemplando campos, eventos e regras de validação

## para apuração do IBS e da CBS

## 01/09/2025 31/10//


## Reforma Tributária - Emenda Constitucional 132/

NT 2024.002 Versão 1. 10

2. Introdução

O Projeto de Lei Complementar Federal PLP 68, aprovado na Câmara dos Deputados e

encaminhado para aprovação junto ao Senado federal, definiu na Seção VIII – Disposições

transitórias, Art. 61, a obrigatoriedade para Estados, o Distrito Federal e os Municípios adaptarem

os sistemas autorizadores de Documentos Fiscais Eletrônicos (DFe) vigentes para utilização de

leiaute padronizado, que permita aos contribuintes informarem os dados relativos ao Imposto sobre

Bens e Serviços (IBS), Contribuição sobre Bens e Serviços (CBS) e Imposto Seletivo (IS).

Como as infraestruturas autorizadoras de Documentos Fiscais Eletrônicos (DFe) das Unidades

Federadas e Municípios, além das aplicações e sistemas dos contribuintes, necessitam de, no

mínimo, 1 ano para o desenvolvimento das alterações necessárias, estamos divulgando esta Nota

Técnica (NT) para implantação, em produção, a partir do dia 31/10/2025, de forma a entrar em

efetiva operacionalização a partir 01/01/2026.

Como as discussões envolvendo a implantação da Reforma Tributária ainda estão em curso,

esclarecemos que esta NT será ajustada ao longo do seu processo de execução, da mesma forma

como ocorre com as demais NT já implementadas.

Esta NT, discutida conjuntamente entre os três entes responsáveis pela implantação da RTC, cria

novos eventos e modifica o leiaute da NF-e e NFC-e, inserindo os grupos e campos opcionais

relacionados à tributação do Imposto sobre Bens e Serviços (IBS), da Contribuição sobre Bens e

Serviços (CBS) e do Imposto Seletivo (IS), em atendimento as alterações previstas na Emenda

Constitucional 132 de 20 de dezembro de 2023 para implementação da Reforma Tributária.


## Reforma Tributária - Emenda Constitucional 132/

NT 2024.002 Versão 1. 10

3. Tipos Básicos da Tributação

Em busca de uma padronização entre os diversos documentos fiscais eletrônicos existentes, esta

NT introduz o arquivo “DFeTiposBasicos_v1.00.xsd” ao conjunto dos arquivos que compõem do

schema de todos os Documentos Fiscais Eletrônicos - DF-e, entre eles a NF-e e NFC-e.

Este arquivo define de forma estruturada a previsão de campos a serem informados para o registro

das informações referentes a tributação do IBS e da CBS em um tipo complexo referenciado no

leiaute padrão da NF-e e NFC-e conforme estrutura demonstrada no item 5, e também será utilizado

nos demais documentos fiscais eletrônicos.

4. Código de Classificação Tributária do IBS/CBS

O grupo de informações do IBS, CBS e IS associado aos itens do documento fiscal contém o Código

de Situação Tributária (CST) e Código de Classificação Tributária (cClassTrib) do IBS, CBS e IS.

O Informe Técnico RT 2024.001 divulga a publicação da tabela com esta codificação, que está

disponível no Portal Nacional da NF-e (www.nfe.fazenda.gov.br), na aba “Documentos”, opção

“Diversos”.

Cada código “cClassTrib” corresponde a um dispositivo específico do PLP68, tornando objetiva a

informação do contribuinte sobre como interpreta que é realizada a tributação do IBS e da CBS para

cada item da NF-e.

A tabela também contém indicadores que vinculam de forma dinâmica códigos “CST-IBS/CBS” e

“cClassTrib” com as Regras de Validação descritas na RT Nota Técnica 2024.002 - IBS/CBS, ou

que contêm informações necessárias para a preparação das apurações assistidas do IBS e da CBS,

em atendimento ao disposto no PLP68.

Caso a versão final aprovada pelo Congresso Nacional altere algum dos dispositivos do PLP68, a

tabela será alterada para refletir esta modificação. Da mesma maneira, a tabela poderá sofrer

alterações em virtude de aperfeiçoamentos, novidades introduzidas em sede de Regulamento, ou

para atender a necessidades relacionadas com a apuração assistida do IBS e da CBS.

5. Finalidades débito e crédito da NF-e

Notas de Débito e Crédito são nomes de instrumentos utilizados mundialmente para documentar

situações contábeis onde é necessário corrigir informações comerciais que foram registradas em

um documento, que no Brasil é a Nota Fiscal.

Esta Nota Técnica cria na NF-e modelo 55 as finalidades de emissão correspondentes. O sentido

das palavras “débito” e “crédito” sempre se referem ao ponto de vista do emissor:

- Uma nota de débito documenta uma situação na qual o emitente registra um aumento no

imposto devido (consequentemente, uma redução no imposto devido pelo adquirente, que é o

destinatário);

- Uma nota de crédito documenta uma situação na qual o emitente registra uma redução no

imposto devido (consequentemente, um aumento no imposto devido pelo adquirente, que é o

destinatário);

As finalidades de emissão “Nota de Ajuste” e “Nota Complementar”, já existentes, são casos

especiais de Nota de Débito; uma nota de entrada emitida para documentar, por exemplo, a


## Reforma Tributária - Emenda Constitucional 132/

NT 2024.002 Versão 1. 10

devolução de mercadoria que havia sido vendida a um consumidor final, é um caso especial de

Nota de Crédito.

A regulamentação do IBS disporá sobre a utilização de notas de crédito e notas de débito para

lançamentos de ajuste, com a finalidade de instrumentalizar a preparação da declaração assistida

a ser oferecida para os contribuintes, de maneira automatizada, a partir de documentos fiscais

eletrônicos, em cumprimento ao que preconiza o PLP 68. A menos que ocorra alteração na

regulamentação do ICMS e do IPI, notas de crédito e notas de débito não poderão ser utilizadas

para ajustes relativos a estes tributos.

6. Alterações no arquivo XML da NF-e

6.1. Esquema gráfico do leiaute do IBS da UF, contemplando IBS, CBS e

Imposto Seletivo.


## Reforma Tributária - Emenda Constitucional 132/

NT 2024.002 Versão 1. 10

6.2. Inserção dos novos campos de informação no leiaute da NF-e e

NFC-e


## Reforma Tributária - Emenda Constitucional 132/

NT 2024.002 Versão 1. 10

6.4. Disposição dos campos do Imposto Seletivo

6.5. Imposto de Bens e Serviços - IBS das UFs e Municípios


## Reforma Tributária - Emenda Constitucional 132/

NT 2024.002 Versão 1. 10

6.6. Disposição dos campos do IBS da UF


## Reforma Tributária - Emenda Constitucional 132/

NT 2024.002 Versão 1. 10

6.7. Disposição dos campos do IBS dos Municípios


## Reforma Tributária - Emenda Constitucional 132/

NT 2024.002 Versão 1. 10

6.9. Disposição dos campos da CBS


## Reforma Tributária - Emenda Constitucional 132/

NT 2024.002 Versão 1. 10

6.10. Disposição dos campos do IBS Monofásico


## Reforma Tributária - Emenda Constitucional 132/

NT 2024.002 Versão 1. 10

6.11. Disposição dos campos totalizadores


## Reforma Tributária - Emenda Constitucional 132/

NT 2024.002 Versão 1. 10

6.12. Indicadores de multa, juros e compra governamental e tipo de nota

de crédito.


## Reforma Tributária - Emenda Constitucional 132/

NT 2024.002 Versão 1. 10

6.13. Referenciamento de Item da NF-e


## Reforma Tributária - Emenda Constitucional 132/

NT 2024.002 Versão 1. 10

7. Protocolo da NFe

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
PR05 tpAmb E PR03 N 1 - 1 1 Identificação do Ambiente:
1=Produção/2=Homologação
PR06 verAplic
E PR03 C 1 - 1 1 - 20 Versão do Aplicativo que processou o Lote.
A versão deve ser iniciada com a sigla da UF nos casos de WS próprio ou a sigla SVAN ou SVRS nos
demais casos.
PR07 chNFe E PR03 N 1 - 1 44 Chave de Acesso da NF-e
PR08 dhRecbto E PR03 D 1 - 1 - Preenchido com a data e hora do processamento (informado também no caso de rejeição).
Formato: “AAAA-MM-DDThh:mm:ssTZD” (UTC – Universal Coordinated Time).
PR09 nProt E PR03 N 0 - 1 15 Número do Protocolo da NF-e, conforme item Erro! Fonte de referência não encontrada.
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

NT 2024.002 Versão 1. 10

8. Leiaute da NF-e (Modelo 55 e 65)

8.1. Grupo B. Identificação da Nota Fiscal eletrônica

# ID Campo Descrição (^) Ele Pai Tipo Ocor. Tam. Observação
16 B12 cMunFG Código do Município de Ocorrência do Fato Gerador do ICMS
E B01 N 1 - 1 7
Informar o município de ocorrência do
fato gerador do ICMS. Utilizar a Utilizar
a Tabela de código de Município do
IBGE
16a B12a cMunFGIBS Código do Município de consumo, fato gerador do IBS / CBS

### E B01 N 0 - 1 7

```
Informar o município de ocorrência do
fato gerador do fato gerador do IBS /
CBS.
Campo preenchido somente quando
“indPres = 5 (Operação presencial, fora
do estabelecimento) ”, e não tiver
endereço do destinatário (Grupo: E05)
ou local de entrega (Grupo: G01).
29 B25 finNFe Finalidade de emissão da NF-e
```
### E B01 N 1 - 1 1

```
1=NF-e normal;
2=NF-e complementar;
3=NF-e de ajuste;
4=Devolução de mercadoria.
5=Nota de crédito
6=Nota de débito
```
```
29e B30 indMultaJuros Indicador de Multa e Juros
E B01 N 0 - 1 1
```
```
0=Indicador de Multa
1=Indicador de Juros
```
29f B31 gCompraGov Grupo de Compra Governamental (^) G B01 0 - 1
29f.1 B32 tpCompraGov Tipo de compra governamental

### E B31 N 1 - 1 1

```
1=União
2=Estados
3=Distrito Federal
4=Municípios
```
```
29f.2 B33 pRedutor Percentual de redução de aliquota em compra governamental
E B31 N 1 - 1 3v2- 4
Conforme Seção XI, Art. 41 PLP
68/2024.
```
29f.3 B34 tipoNotaCredito Indicador do tipo de nota de crédito que está sendo utilizada (^) E B01 N 0 - 1 1 Aguardando GT06. A definir.


## Reforma Tributária - Emenda Constitucional 132/

NT 2024.002 Versão 1. 10

8.2. Grupo N01. ICMS Normal e ST

# ID Campo Descrição (^) Ele Pai Tipo Ocor. Tam. Observação
164 N01 ICMS Informações do ICMS da Operação própria e ST
CG M01 - 0 - 1 -
Informar apenas um dos grupos de tributação do
ICMS (ICMS00, ICMS10, ...) (v2.0)
8.3. Grupo UB. Informações dos tributos IBS / CBS e Imposto Seletivo
# ID Campo Descrição (^) Ele Pai Tipo Ocor. Tam. Observação
324.01 UB01 IBSCBSSel Informações do Imposto de Bens e Serviços -
IBS, Contribuição de Bens e Serviços - CBS e
Imposto Seletivo

### G H01 - 0 - 1 -

324.02 UB02 seletivo Informações do Imposto Seletivo (^) G UB01 - 0 - 1 -
324.03 UB03 CST Código de Situação Tributária do Imposto Seletivo
E UB02 N 1 - 1 3
Utilizar tabela CÓDIGO DE CLASSIFICAÇÃO
TRIBUTÁRIA DO IMPOSTO SELETIVO
324.04 UB04 cClassTrib Código de Classificação Tributária do Imposto
Seletivo E^ UB02^ C^1 -^1 
Utilizar tabela CÓDIGO DE CLASSIFICAÇÃO
TRIBUTÁRIA DO IMPOSTO SELETIVO
324.05 UB05 gImpSel Grupo de Informações do Imposto Seletivo (^) G UB02 - 0 - 1 -
324.06 UB06 vBCImpSel Valor da Base de Cálculo do Imposto Seletivo (^) E UB05 N 1 - 1 13v
324.07 UB07 pImpSel Alíquota do Imposto Seletivo (^) E UB05 N 1 - 1 3v2- 4
324.08 UB08 pImpSelEspec Alíquota específica por unidade de medida
apropriada
E UB05 N 0 - 1 3v2- 4
324.09 UB09 uTrib Unidade de Medida Tributável (^) E UB05 C 0 - 1 1 - 6
324.10 UB10 qTrib Quantidade Tributável (^) E UB05 N 0 - 1 11v0- 4
324.11 UB11 vImpSel Valor do Imposto Seletivo (^) E UB05 N 1 - 1 13v
324.12 UB12 CST Código de Situação Tributária do IBS e CBS
E UB01 N 0 - 1 3
Utilizar tabela CÓDIGO DE CLASSIFICAÇÃO
TRIBUTÁRIA DO IBS E DA CBS
324.13 UB13 cClassTrib Código de Classificação Tributária do IBS e CBS
E UB01 C 0 - 1 6
Utilizar tabela CÓDIGO DE CLASSIFICAÇÃO
TRIBUTÁRIA DO IBS E DA CBS
324.14 UB14 indPerecimento Indicador de Perecimento, furto, roubo ou extravio
E UB01 C 0 - 1 1
“1”= Perecimento, roubo, furto ou extravio do
item
324.15 UB15 gIBSCBS Grupo de Informações do IBS, CBS e Imposto
Seletivo

### CG UB01 1 - 1

324.16 UB16 vBC Base de cálculo do IBS e CBS (^) E UB15 N 1 - 1 3v2- 4


## Reforma Tributária - Emenda Constitucional 132/

NT 2024.002 Versão 1. 10

# ID Campo Descrição (^) Ele Pai Tipo Ocor. Tam. Observação
324.17 UB17 gIBSUF Grupo de Informações do IBS para a UF (^) G UB15 - 1 - 1 -
324.18 UB18 pIBSUF Alíquota do IBS de competência das UF (^) E UB17 N 1 - 1 3v2- 4 Alíquota vigente do IBS da UF
324.19 UB19 - x- Sequencia XML (^) G UB17 - 0 - 1 -
324.20 UB20 vTribOP Valor bruto do tributo na operação
E UB19 N 1 - 1 13v
Valor do tributo considerando BC x Alq do IBS,
sem considerar qualquer desoneração.
324.21 UB21 gDif Grupo de Informações do Diferimento (^) G UB19 0 - 1
324.22 UB22 pDif Percentual do diferimento (^) E UB21 N 1 - 1 3v2- 4
324.23 UB23 vDif Valor do Diferimento (^) E UB21 N 1 - 1 13v
324.24 UB24 gDevTrib Grupo de Informações da devolução de tributos (^) G UB19 0 - 1
324.25 UB25 vDevTrib Valor do tributo devolvido
E UB24 N 1 - 1 13v
Valor do tributo devolvido. No fornecimento de
energia elétrica, água, esgoto e gás natural e
em outras hipóteses definidas no regulamento
324.26 UB26 gRed Grupo de informações da redução da alíquota (^) G UB17 - 0 - 1 -
324.27 UB27 pRedAliq Percentual da redução de alíquota E UB26 N 1 - 1 3v2- 4
324.28 UB28 pAliqEfet Aliquota Efetiva do IBS de competência das UF
que será aplicada a Base de Cálculo
E UB26 N 1 - 1 3v2- 4
Alíquota efetiva, após aplicação da redução de
alíquota.
324.29 UB29 gDeson Grupo de informações da Desoneração
G UB17 - 0 - 1 -
Grupo de informações da Desoneração.
Exemplo 1: Art. 442, §4. Operações com ZFM
e ALC. Exemplo 2: Operações com
suspensão do tributo.
324.30 UB30 CST Código de Situação Tributária do IBS e CBS
E UB29 N 0 - 1 3
Informado como se a operação fosse tributada
integralmente. Utilizar tabela CÓDIGO DE
CLASSIFICAÇÃO TRIBUTÁRIA DO IBS E DA
CBS
324.31 UB31 cClassTrib Código de Classificação Tributária do IBS e CBS
E UB29 C 0 - 1 6
Informado como se a operação fosse tributada
integralmente. Utilizar tabela CÓDIGO DE
CLASSIFICAÇÃO TRIBUTÁRIA DO IBS E DA
CBS
324.32 UB32 vBC Valor da BC
E UB29 N 0 - 1 13v
Info: Avaliando retirar e deixar a BC do grupo
anterior
324.33 UB33 pAliq Valor da alíquota
E UB29 N 1 - 1 3v2- 4
Informado como se a operação fosse tributada
integralmente
324.34 UB34 vDeson Valor desonerado E UB29 N 1 - 1 13v
324.35 UB35 vIBSUF Valor do IBS de competência da UF E UB17 N 1 - 1 13v
324.36 UB36 gIBSMun Grupo de Informações do IBS para o município G UB15 - 1 - 1 -
324.37 UB37 pIBSMun Alíquota do IBS de competência dos Municípios (^) E UB36 N 1 - 1 3v2- 4 Alíquota vigente do IBS do Município
324.38 UB38 - x- Sequencia XML (^) G UB41 - 0 - 1 -


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

# ID Campo Descrição (^) Ele Pai Tipo Ocor. Tam. Observação
324.39 UB39 vTribOP Valor bruto do tributo na operação
E UB38 N 1 - 1 13v2
Valor do tributo considerando BC x Alq do IBS,
sem considerar qualquer desoneração.
324.40 UB40 gDif Grupo de Informações do Diferimento (^) G UB38 0 - 1
324.41 UB41 pDif Percentual do diferimento (^) E UB40 N 1 - 1 3v2- 4
324.42 UB42 vDif Valor do Diferimento (^) E UB40 N 1 - 1 13v2
324.43 UB43 gDevTrib Grupo de Informações da devolução de tributos (^) G UB38 0 - 1
324.44 UB44 vDevTrib Valor do tributo devolvido
E UB43 N 1 - 1 13v2
Valor do tributo devolvido. No fornecimento de
energia elétrica, água, esgoto e gás natural e
em outras hipóteses definidas no regulamento
324.45 UB45 gRed Grupo de informações da redução da alíquota (^) G UB36 - 0 - 1 -
324.46 UB46 pRedAliq Percentual da redução de alíquota E UB45 N 1 - 1 3v2- 4
324.47 UB47 pAliqEfet Aliquota Efetiva do IBS de competência dos
Municípios que será aplicada a Base de Cálculo
E UB45 N 1 - 1 3v2- 4
Alíquota efetiva, após aplicação da redução de
alíquota.
324.48 UB48 gDeson Grupo de informações da Desoneração
G UB36 - 0 - 1 -
Grupo de informações da Desoneração.
Exemplo 1: Art. 442, §4. Operações com ZFM
e ALC. Exemplo 2: Operações com
suspensão do tributo.
324.49 UB49 CST Código de Situação Tributária do IBS e CBS
E UB48 N 0 - 1 3
Informado como se a operação fosse tributada
integralmente. Utilizar tabela CÓDIGO DE
CLASSIFICAÇÃO TRIBUTÁRIA DO IBS E DA
CBS
324.50 UB50 cClassTrib Código de Classificação Tributária do IBS e CBS
E UB48 C 0 - 1 6
Informado como se a operação fosse tributada
integralmente. Utilizar tabela CÓDIGO DE
CLASSIFICAÇÃO TRIBUTÁRIA DO IBS E DA
CBS
324.51 UB51 vBC Valor da BC
E UB48 N 1 - 1 13v2
Info: Avaliando retirar e deixar a BC do grupo
anterior
324.52 UB52 pAliq Valor da alíquota
E UB48 N 1 - 1 3v2- 4
Informado como se a operação fosse tributada
integralmente
324.53 UB53 vDeson Valor desonerado E UB48 N 1 - 1 13v2
324.54 UB54 vIBSMun Valor do IBS de competência dos Municípios (^) E UB36 N 1 - 1 13v2


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

# ID Campo Descrição (^) Ele Pai Tipo Ocor. Tam. Observação
324.55 UB55 gIBSCredPres Grupo de Informações do Crédito Presumido
referente ao IBS

### G UB15 - 0 - 1 -

```
Grupo de Informações do Crédito Presumido
do IBS, quando aproveitado pelo emitente do
documento. Exemplos: 1 - Aquisição de PR
não contribuinte. 2 - Tomador de serviço de
transporte de TAC PF não contrib. 3 -
Aquisição de pessoa física com destino a
reciclagem. 4 - Aquisição de bens móveis de
PF não contrib. para revenda (veículos /
brecho). 5 - Regime opcional para
cooperativa.
324.56 UB56 cCredPres Código de Classificação do Crédito Presumido
E UB55 N 1 - 1 2
Utilizar tabela CÓDIGO DE CLASSIFICAÇÃO
DO CRÉDITO PRESUMIDO
324.57 UB57 pCredPres Percentual do Crédito Presumido E UB55 N 1 - 1 3v2- 4
324.58 UB58 vCredPres Valor do Crédito Presumido E UB55 CE 1 - 1 13v2
324.59 UB59 vCredPresCondSus Valor do Crédito Presumido em condição
suspensiva.
```
```
E UB55 CE 1 - 1 13v2
```
```
Valor do Crédito Presumido Condição
Suspensiva. Preencher apenas para
cClassCredPres com indicação de Condição
Suspensiva.
Incluir regra de validação para permitir apenas
com o código 4 - Aquisição de bens móveis de
PF não contrib. para revenda (veículos /
brecho).
```
324.60 UB60 gCBS Grupo de Informações da CBS (^) G UB15 - 1 - 1 -
324.61 UB61 pCBS Alíquota da CBS (^) E UB60 N 1 - 1 3v2- 4 Alíquota vigente da CBS.
324.62 UB62 - x- Sequencia XML (^) G UB60 - 0 - 1 -
324.63 UB63 vTribOp Valor bruto do tributo na operação
E UB62 N 1 - 1
Valor do tributo considerando BC x Alq da CBS,
sem considerar qualquer desoneração.
324.64 UB64 gCBSCredPres
Grupo de Informações do Crédito Presumido
referente a CBS

### G UB62 - 0 - 1

```
Grupo de Informações do Crédito Presumido
da CBS, quando aproveitado pelo emitente
do documento. Exemplos: 1 - Aquisição de
PR não contribuinte. 2 - Tomador de serviço
de transporte de TAC PF não contrib. 3 -
Aquisição de pessoa física com destino a
reciclagem. 4 - Aquisição de bens móveis de
PF não contrib. para revenda (veículos /
brecho). 5 - Regime opcional para
cooperativa.
324.65 UB65 cCredPres Código de Classificação do Crédito Presumido
E UB64 N 1 - 1 2
Utilizar tabela CÓDIGO DE CLASSIFICAÇÃO
DO CRÉDITO PRESUMIDO
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

# ID Campo Descrição (^) Ele Pai Tipo Ocor. Tam. Observação
324.66 UB66 pCredPres Percentual do Crédito Presumido (^) E UB64 N 1 - 1 3v2- 4
324.67 UB67 vCredPres Valor do Crédito Presumido (^) E UB64 CE 1 - 1 13v2
324.68 UB68 vCredPresCondSus Valor do Crédito Presumido em condição
suspensiva.
E UB64 CE 1 - 1 13v2
Valor do Crédito Presumido Condição
Suspensiva. Preencher apenas para
cClassCredPres com indicação de Condição
Suspensiva.
324.69 UB69 gDif Grupo de Informações do Diferimento (^) G UB62 - 0 - 1
324.70 UB70 pDif Percentual do diferimento E UB69 N 1 - 1 3v2- 4
324.71 UB71 vDif Valor do Diferimento E UB69 N 1 - 1 13v2
324.72 UB72 gDevTrib Grupo de Informações da devolução de tributos G UB62 - 0 - 1
324.73 UB73 vDevTrib Valor do tributo devolvido
E UB72 N 1 - 1 13v2
Valor do tributo devolvido. No fornecimento de
energia elétrica, água, esgoto e gás natural e
em outras hipóteses definidas no regulamento
324.74 UB74 gRed Grupo de informações da redução da alíquota G UB60 - 0 - 1 -
324.75 UB75 pRedAliq Percentual da redução de alíquota (^) E UB74 N 1 - 1 3v2- 4
324.76 UB76 pAliqEfet Alíquota Efetiva da CBS que será aplicada a Base
de Cálculo
E UB74 N 1 - 1 3v2- 4
Alíquota efetiva, após aplicação da redução de
alíquota.
324.77 UB77 gDeson Grupo de informações da Desoneração
G

### UB60

### - 0 - 1 -

```
Grupo de informações da Desoneração.
Exemplo 1: Art. 442, §4. Operações com ZFM
e ALC. Exemplo 2: Operações com
suspensão do tributo.
324.78 UB78 CST Código de Situação Tributária do IBS e CBS
```
```
E UB77 N 0 - 1 3
```
```
Informado como se a operação fosse tributada
integralmente. Utilizar tabela CÓDIGO DE
CLASSIFICAÇÃO TRIBUTÁRIA DO IBS E DA
CBS
324.79 UB79 cClassTrib Código de Classificação Tributária do IBS e CBS
```
```
E UB77 C 0 - 1 6
```
```
Informado como se a operação fosse tributada
integralmente. Utilizar tabela CÓDIGO DE
CLASSIFICAÇÃO TRIBUTÁRIA DO IBS E DA
CBS
324.80 UB80 vBC Valor da BC
E UB77 N 1 - 1 13v2
Info: Avaliando retirar e deixar a BC do grupo
anterior
324.81 UB81 pAliq Valor da alíquota
E UB77 N 1 - 1 3v2- 4
Informado como se a operação fosse tributada
integralmente
```
324.82 UB82 vDeson Valor desonerado (^) E UB77 N 1 - 1 13v2
324.83 UB83 vCBS Valor da CBS (^) E UB60 N 1 - 1 13v2
324.84 UB84 gIBSCBSMono Grupo de Informações do IBS e CBS em
operações com imposto monofásico

### CG UB01 1 - 1

```
Monofasia dos Combustíveis
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

# ID Campo Descrição (^) Ele Pai Tipo Ocor. Tam. Observação
324.85 UB85 qBCMono Quantidade tributada na monofasia
E UB84 N 0 - 1 11v0- 4
Informar a BC quantidade conforme unidade de
medida estabelecida na legislação para o
produto.
324.86 UB86 adRemIBS Alíquota ad rem do IBS (^) E UB84 N 1 - 1 3v2- 4
324.87 UB87 adRemCBS Alíqutoa ad rem da CBS (^) E UB84 N 1 - 1 3v2- 4
324.88 UB88 vIBSMono Valor do IBS monofásico
E UB84 N 1 - 1 13v2
O valor do imposto é obtido pela multiplicação
da alíquota ad rem pela quantidade do produto
conforme unidade de medida estabelecida na
legislação.
324.89 UB89 vCBSMono Valor da CBS monofásica
E UB84 N 1 - 1 13v2
O valor do imposto é obtido pela multiplicação
da alíquota ad rem pela quantidade do produto
conforme unidade de medida estabelecida na
legislação.
324.90 UB90 - x- Sequencia XML
G

### UB84

### 0 - 1

```
Uso em operações com combustíveis
derivados de petróleo (Gasolina A) [ou *Óleo
Diesel A*] para retenção do imposto sobre o
biocombustível a ser misturado. Art 173 PLP
68/24.
324.91 UB91 qBCMonoReten Quantidade tributada sujeita à retenção na
monofasia E UB90 N 0 - 1 11v0- 4
```
```
Informar a BC do ICMS sujeita a retenção em
quantidade conforme unidade de medida
estabelecida na legislação para o produto.
324.92 UB92 adRemIBSREten Alíquota ad rem do imposto sujeito a retenção E UB90 N 1 - 1 3v2- 4
324.93 UB93 vIBSMonoReten Valor do IBS monofásico sujeito a retenção
E UB90 N 1 - 1 13v2
```
```
Valor do IBS com retenção, a ser somado ao
valor de IBS a ser recolhido.
324.94 UB94 - x- Sequencia XML
```
```
G UB84 0 - 1
```
```
Uso em operações com crédito presumido.
Exemplo: compra de combustível por
transportador de passageiros, venda para
órgão público, venda equiparada a
exportação, entre outros.
324.95 UB95 pCredPresIBS Percentual de crédito presumido do IBS
monofásico.
E UB94 N 1 - 1 3v2- 4
```
324.96 UB96 vCRedPresIBS Valor do crédito presumido do IBS monofásico. (^) E UB94 N 1 - 1 13v2
324.97 UB97 pCredPresCBS Percentual de crédito presumido da CBS
monofásica.
E UB94 N 1 - 1 3v2- 4
324.98 UB98 vCredPresCBS Valor do crédito presumido da CBS monofásica. (^) E UB94 N 1 - 1 13v2
324.99 UB99 - x- Sequencia XML
G UB84 0 - 1
Operações com diferimento, aplicado aos
biocombustíveis. Exemplo: operação do
produtor de biocombustível (usina).
324.100 UB100 pDifIBS Percentual do diferimento do imposto monofásico. E UB99 N 1 - 1 3v2- 4 A ser aplicado em vIBSMono.


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

# ID Campo Descrição (^) Ele Pai Tipo Ocor. Tam. Observação
324.101 UB101 vIBSMonoDif Valor do IBS mono diferido. (^) E UB99 N 1 - 1 13v2 A ser deduzido do valor do IBS.
324.102 UB102 pDifCBS Percentual do diferimento do imposto monofásico (^) E UB99 N 1 - 1 3v2- 4 A ser aplicado em vCBSMono
324.103 UB103 vCBSMonoDif Valor do CBS Mono diferido. (^) E UB99 N 1 - 1 13v2 A ser deduzido do valor da CBS.
324.104 UB104 vTotIBSMono Total de IBS Monofásico.
E UB84 N 1 - 1 13v2
Considerando a adição do valor de retenção e
deduzindo a parcela diferida conforme o caso.
324.105 UB105 vTotCBSMono Total da CBS Monofásica.
E UB84 N 1 - 1 13v2
Considerando a adição do valor de retenção e
deduzindo a parcela diferida conforme o caso.
8.4. Grupo VB. Referenciamento de item de outro Documento Fiscal Eletrônico - DF-e
# ID Campo Descrição (^) Ele Pai Tipo Ocor. Tam. Observação
325h VB01 DFeReferenciado Documento Fiscal Eletrônico Referenciado
G H01 0 - 1 1
Grupo para referenciamento de itens de outro
DF-e.
325i VB02 chaveAcesso Chave de acesso do DF-e referenciado (^) E VB01 N 1 - 1 44 Chave de acesso do DF-e referenciado.
325j VB03 nItem Número do item do documento referenciado.
E VB01 N 1 - 1 3
Corresponde ao atributo “nItem” do elemento “det”
do documento original.
Se o documento referenciado não tiver item, indicar
“1”
8.5. Grupo W03. Total da NF-e - IBS / CBS / IS
# ID Campo Descrição (^) Ele Pai Tipo Ocor. Tam. Observação
355.1 W31 IBSCBSSelTot Totais da NF-e com IBS, CBS e IS

### G A01 - 0 - 1 -

```
O grupo de valores totais da NF-e deve ser
informado com o somatório do campo
correspondente dos itens.
```
```
O IBS, a CBS e o IS são por fora, por isso
seus valores devem ser adicionados ao valor
total da NF.
```
355.2 W32 gSel Grupo total do imposto seletivo (^) G W31 - 0 - 1 -
355.3 W33 vBCSel Total da base de cálculo do imposto seletivo (^) E W32 N 1 - 1 11v0- 4


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

# ID Campo Descrição (^) Ele Pai Tipo Ocor. Tam. Observação
355.4 W34 vImpSel Total do imposto seletivo (^) E W32 N 1 - 1 13v2
355.5 W35 vBCIBSCBS Valor total da BC do IBS e da CBS (^) E W31 N 1 - 1 13v2
355.6 W36 gIBS Grupo total do IBS G W31 - 1 - 1 -
355.7 W37 gIBSUFTot Grupo total do IBS da UF (^) G W36 - 1 - 1 -
355.8 W38 vDif Valor total do diferimento E W37 N 1 - 1 13v2
355.9 W39 vDevTrib Valor total de devolução de tributos (^) E W37 N 1 - 1 13v2
355.10 W40 vDeson Valor total de desoneração E W37 N 1 - 1 13v2
355.11 W41 vIBSUF Valor total do IBS da UF (^) E W37 N 1 - 1 13v2
355.12 W42 gIBSMunTot Grupo total do IBS do Município G W36 - 1 - 1 -
355.13 W43 vDif Valor total do diferimento (^) E W42 N 1 - 1 13v2
355.14 W44 vDevTrib Valor total de devolução de tributos E W42 N 1 - 1 13v2
355.15 W45 vDeson Valor total de desoneração (^) E W42 N 1 - 1 13v2
355.16 W46 vIBSMun Valor total do IBS do Município E W42 N 1 - 1 13v2
355.17 W47 vIBSTot Valor total do IBS
E W42 N 1 - 1 13v2
355.18 W48 vCresPres Valor total do crédito presumido
E W36 N 1 - 1 13v2
355.19 W49 vCredPresCondSus Valor total do crédito presumido em condição suspensiva.
E W36 N 1 - 4 13v2
355.20 W50 gCBS Grupo total da CBS
G W31 - 1 - 1 -
355.21 W51 vCresPres Valor total do crédito presumido
E W49 N 1 - 1 13v2
355.22 W52 vCredPresCondSus Valor total do crédito presumido em condição suspensiva.
E W49 N 1 - 4 13v2
355.23 W53 vDif Valor total do diferimento
E W49 N 1 - 1 13v2
355.24 W54 vDevTrib Valor total de devolução de tributos
E W49 N 1 - 1 13v2


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

# ID Campo Descrição (^) Ele Pai Tipo Ocor. Tam. Observação
355.25 W55 vDeson Valor total de desoneração
E W49 N 1 - 1 13v2
355.26 W56 vCBS Valor total da CBS
E W49 N 1 - 1 13v2

### 355.27

```
W57 gMono Grupo total da Monofasia G W31 - 0 - 1 -
```
### 355.28

```
W58 vTotIBSMono Total do IBS monofásico E W57 N 1 - 1 13v2
```
### 355.29

W59 (^) vTotCBSMono Total da CBS monofásica E W57 N 1 - 1 13v2

### 355.30

W60 vTotNF Valor total da NF-e com IBS / CBS / IS E W57 N 1 - 1 13v2

9. Regras de Validação

9.1. Grupo B. Identificação da Nota Fiscal eletrônica

```
Campo-
Seq
Modelo Regra de Validação Aplic. Msg Descrição Erro
```
```
B12a- 20 65
Se infomado município do fato gerador IBS (tag: cMunUFIBS) , validar se
operação é presencial fora do estabelecimento (tag: indPres = 5)
Obrig. 1000
```
```
Rejeição: Município do fato gerador do IBS deve ser
informado apenas em operação presencial fora do
estabelecimento.
```
### B25- 60 55

```
Se NF-e complementar ou de crédito (tag:finNFe=2 ou 5):
```
- UF da NF-e referenciada diferente da UF do emitente (NF-e, NFC-e,
NF modelo 1) (NT 2013/003)

```
Obrig. 678
Rejeição: NF referenciada com UF diferente da NF-e
complementar
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

```
Campo-
Seq
Modelo Regra de Validação Aplic. Msg Descrição Erro
```
### B25- 80 55/65

```
Se finalidade da NF-e igual a crédito ou débito (tag:finNFe=5 ou 6):
Não pode ser informado ICMS (tag: ICMS), ISSQN (tag: ISSQN), IPI
(tag: IPI), II (tag: II), PIS (tag: PIS), PIS ST (tag: PISST), COFINS(tag:
COFINS), COFINS ST (tag: COFINSST), ICMS UF Destino (tag:
ICMSUFDest) e Imposto Devolvido (tag: impostoDevol).
```
```
Obrig. 1001
```
```
Rejeição: NF-e com finalidade de débito ou crédito somente
para IBS/CBS.
```
### B25- 90 55/65

```
Se finalidade da NF-e diferente de crédito ou débito (tag:finNFe=5 ou 6):
Deve ser informado ICMS (tag: ICMS) ou ISSQN (tag: ISSQN).
Obrig. 1002 Rejeição: NF-e sem informação de ICMS / ISSQN.
```
### B25- 100 55/65

```
Se NF-e de crédito (tag: finNF-e = 5), somente pode ser referenciada
NF-e (tag: NFref/refNfe) e modelo deve ser 55.
Obrig. 1003
Rejeição: NF-e de cŕedito referencia documento fiscal
diferente de NF-e modelo 55.
```
### B25- 110 55

```
Se NF-e de devolução de mercadoria (tag: finNF-e = 4 ) exige
referenciamento de NF-e (tag: refNFe) do item da NF-e (tag:
DFeReferenciado/nItem).
```
```
Obrig. 1102
```
```
Rejeição: NF-e de devolução de mercadoria exige
referenciamento do item da NF-e original.
```
```
B25b- 50 65
NFC-e com entrega a domicílio (tag:indPres=4) e endereço do
destinatário não foi informado (id: E5, grupo: enderDest)
Obrig. 1004
Rejeição: NFC-e de entrega a domicílio e não informado
endereço do destinatário.
```
```
B25b- 60 65
NFCe com operação presencial fora do estabelecimento (tag: indPres =
5) e município do fato gerador do IBS não infomado (tag: cMunFGIBS)
Obrig. 1005
Rejeição: Operação presencial fora do estabelecimento deve
informar o município do fato gerador do IBS.
```
### B30- 10 55

```
Se informado indicador de juros e multa (tag: indMultaJuros), validar se é
nota de crédito (tag: finNFe = 5) ou nota de débito (tag: finNFe = 6)
Obrig. 1006
Rejeição: Indicador de multa e juros deve ser informado
somente para nota de crédito ou débito
```
### B30- 20 55

```
Indicador de juros e multa (tag: indMultaJuros) exige referenciamento de
NF-e (tag: refNFe) do item da NF-e (tag: DFeReferenciado/nItem).
Obrig. 1007
Rejeição: NF-e com indicador de multa e juros exige
referenciamento do item da NF-e original.
```
### B32- 10 55

```
Se informado tpCompraGov (nota de compra governamental), verificar
se alíquota de outros entes é diferente a zero.
Para compra da união (tag: tpCompraGov = 1)
Alíquota do IBS da UF (tag: pUBSUF) deve ser igual a zero.
Alíquota do IBS do Município (tag: pUBSMun) deve ser igual a zero.
Para compra estadual (tag: tpCompraGov = 2)
Alíquota do IBS do Município (tag: pUBSMun) deve ser igual a zero.
Alíquota da CBS (tag: pCBS) deve ser igual a zero.
Para compra do DF (tag: tpCompraGov = 3)
Alíquota da CBS (tag: pCBS) deve ser igual a zero.
Para compra municipal (tag: tpCompraGov = 4)
Alíquota do IBS da UF (tag: pUBSUF) deve ser igual a zero.
Alíquota da CBS (tag: pCBS) deve ser igual a zero.
Observação: Conforme PLP 68/24, Art 41, § 1o e seus Incisos I, II, III e
IV
```
```
Obrig. 1008
```
```
Rejeição: Nota de compra governamental e alíquota dos
outros entes diferente de zero.
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

```
Campo-
Seq
Modelo Regra de Validação Aplic. Msg Descrição Erro
```
### B34- 10 55

```
Se informado tipo de nota de cŕedito (tag: tipoNotaCredito), validar se
finalidade = 5 (tag:finNfe) - Nota de cŕedito
Obrig. 1009
Rejeição: Tipo de nota de crédito deve ser informado apenas
para nota com finalidade de nota de cŕedito.
```
9.2. Grupo BA. Documento Fiscal Referenciado

```
Campo-
Seq
```
```
Modelo Regra de Validação Aplic. Msg Descrição Erro
```
### BA01- 20 55/65

```
Não pode haver documento referenciado (tag: NFref), quando existe
referenciamento de item (tag: DFeReferenciado/nItem)
Obrig. 1010
```
```
Rejeição: NF-e com referenciamento de documento e de
item de documento.
```
9.3. Grupo UB. Informações dos tributos IBS / CBS e Imposto Seletivo

```
Campo-
Seq
Modelo Regra de Validação Aplic. Msg Descrição Erro
```
### UB02- 10 55/65

```
Não é permitido uso do Imposto Seletivo (grupo: seletivo) para este
cClass.
Obrig. 1011
```
```
Rejeição: Não é permitido uso do Imposto Seletivo para
esta classificação da operação.
```
```
UB02- 20 55/65 É exigido uso do Imposto Seletivo (grupo: seletivo) para este cClass. Obrig. 1012
Rejeição: É exigido o uso do Imposto Seletivo para esta
classificação da operação.
```
```
UB02- 30 55/65 É exigido uso do Imposto Seletivo (grupo: seletivo) para este NCM. Obrig. 1013
Rejeição: É exigido o uso do Imposto Seletivo para esta
classificação da operação para este NCM.
```
### UB03- 10 55/65

```
Se CST do Imposto Seletivo for informado, este deve existir na tabela de
Código de Situação Tributária (tag: seletivo/CST).
Obrig. 1014 Rejeição: CST do Imposto Seletivo informado inexistente
```
### UB04- 10 55/65

```
Se cClassTrib for informado, este deve existir na tabela de Classificação
Tributária do Imposto Seletivo (tag: seletivo/cClassTrib)
Obrig. 1015
Rejeição: Classificação Tributária do Imposto Seletivo
informada inexistente
```
### UB07- 10 55/65

```
Se cClassTrib informado exigir grupo do Imposto Seletivo e exigir a
alíquota (tag: pImpSel) diferente de Zero e alíquota (tag: pImpSel) do
imposto seletivo for zero. (tag: seletivo/pImpSel).
```
```
Obrig. 1016
Rejeição: CST/cClassTrib do Imposto Seletivo obriga
informação de alíquota de Imposto Seletivo
```
### UB08- 10 55 - 65

```
Se cClassTrib informado exigir grupo do Imposto Seletivo e exigir a
alíquota específica (tag: gImpSel/pImpSelEspec) diferente de Zero e
NCM do item for 2401, 2402, 2403, 2404, 2203, 2204, 2205, 2206, 2208.
```
```
Obrig. 1017
Rejeição: Obrigatório informação de alíquota específica de
Imposto Seletivo
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

```
Campo-
Seq
Modelo Regra de Validação Aplic. Msg Descrição Erro
```
### UB09- 10 55/65

```
Se cClassTrib informado exigir grupo do Imposto Seletivo (grupo:
seletivo) e unidade tributável (tag: seletivo/uTrib) e quantidade tributável
(tag: seletivo/qtrib) não informadas ou iguais a zero.
```
```
Obrig. 1018
Rejeição: Unidade tributável (UTrib) e Quantidade tributável
(QTrib) do imposto seletivo não informados.
```
### UB11- 10 55/65

```
Se informado imposto seletivo (tag: imposto/seletivo):
```
```
Valor do IS (vImpSel) = BC (tag: seletivo/vBC) * Alq (tag:
seletivo/pImpSel)
```
```
Observação 1:
Se informada alíquota específica (tag: seletivo/pImpSelEspec):
Valor do IS (vImpSel) = qtd (tag: seletivo/vTrib) * Alq (tag:
seletivo/pImpSelEspec)
```
```
Observação 2: Aceitar uma tolerância de 0,01 a mais ou a menos.
```
```
Obrig. 1019
Rejeição: Valor do Imposto Seletivo diferente de Base de
Cálculo x Alíquota
```
### UB12- 10 55/65

```
Se CST do IBS/CBS for informado, este deve existir na tabela de Código
de Situação Tributária (tag: gIBSCBS/CST)
Obrig. 1020 Rejeição: CST do IBS/CBS informado inexistente
```
### UB12- 20 55/65

```
Se o CST do IBS/CBS (tag: gIBSCBS/CST) informado VEDAR
preenchimento do grupo de informações específicas do IBS/CBS, este
grupo NÃO DEVE estar informado (grupo: imposto/IBSCBSSel)
```
```
Obrig. 1021
Rejeição: Grupo IBS/CBS não deve ser preenchido para o
CST informado
```
### UB12- 30 55/65

```
Se o CST do IBS/CBS (tag: gIBSCBS/CST) informado EXIGIR
preenchimento do grupo de informações específicas do IBS/CBS, este
grupo DEVE estar informado (grupo: imposto/IBSCBSSel)
```
```
Obrig. 1022
```
```
Rejeição: Grupo IBS/CBS deve ser preenchido para o CST
informado
```
### UB13- 10 55/65

```
Se cClassTrib for informado, este deve existir na tabela de Classificação
Tributária do IBS/CBS (tag: gIBSCBS/cClassTrib)
Obrig. 1023
```
```
Rejeição: Classificação Tributária do IBS/CBS informada
inexistente
```
```
UB13- 20 55/65
cClassTrib (tag: gIBSCBS/cClassTrib) for informado deve ser compatível
com CST (tag: gIBSCBS/CST)
Obrig. 1024
Rejeição: Rejeição: Classificação Tributária incompatível
com o CST informado
```
### UB14- 10 55/65

```
Indicador de perecimento (tag: indPerecimento = 1) apenas para NF-e
modelo 55 e tipo de operação de saída (tag: tpNF = 1) e CNPJ / CPF do
emitente igual ao CNPJ / CPF do destinatário.
```
```
Obrig. 1025
```
```
Rejeição: Indicador de perecimento deve ser infomado
apenas para NF-e operação de saída para o próprio
contribuinte.
```
```
UB18- 10 55/65
Alíquota do IBS da UF (tag: pIBSUF) deve ser igual a 0,1% para
documento com data de emissão no ano de 2026. Art 342 PL 68/24.
Obrig. 1026
Rejeição: Alíiquota do IBS da UF deve ser igual a 0,1%
para documento emitido em 2026
```
### UB18- 20 55/65

```
Alíiquota do IBS da UF (tag: pIBSUF) deve ser igual a 0,05% para
documento com data de emissão nos anos de 2027 e 2028. Art 343 PL
68/24.
```
```
Obrig. 1027
Rejeição: Alíiquota do IBS da UF deve ser igual a 0,05%
para documento emitido em 2027 e 2028
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

```
Campo-
Seq
Modelo Regra de Validação Aplic. Msg Descrição Erro
```
### UB20- 10 55/65

```
Se informado grupo IBS de competência das Unidades Federadas
(gIBSUF)
O valor bruto do tributo do IBS das UF deve ser igual à multiplicação da
Base de Cálculo pela Alíquota.
```
```
vTribOP (tag: gIBSUF/vTribOP) = vBC (tag: gIBSCBS/vBC) * pIBSUF
(tag: gIBSUF/pIBSUF)
```
```
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos.
```
```
Obrig. 1028
Rejeição: NF-e com valor bruto do tributo do IBS das UFs
calculado incorretamente.
```
```
UB22- 10 55/65 Não é permitido o uso de diferimento (grupo: gDif) para este cClassTrib Obrig. 1029
Rejeição: Não é permitido o uso de Diferimento para esta
classificação da operação.
```
```
UB22- 20 55/65 É exigido o uso de diferimento da UF (grupo: gDif) para este cClassTrib Obrig. 1030
Rejeição: É exigido o uso de Diferimento da UF para esta
classificação da operação.
```
### UB23- 10 55/65

```
Se informado grupo do Diferimento (gIBSUF/gDif):
O valor do Diferimento (vDif) deverá ser resultante da Base de Cálculo x
Percentual do Diferimento (vBC x pDif)
```
```
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos
```
```
Obrig. 1031
Rejeição: Valor do Diferimento da UF diferente de Base de
Cálculo x Percentual
```
### UB26- 10 55/65

```
Não é permitido o uso de redução de alíquota (grupo: gRed) para este
cClassTrib.
```
```
Obrig. 1032
Rejeição: Não é permitido o uso de Redução de Alíquota
para esta classificação da operação.
```
```
UB26- 20 55/65
```
```
É exigido o uso de redução de alíquota (grupo: gRed) para este
cClassTrib
Obrig. 1033
```
```
Rejeição: É exigido o uso de Redução de Alíquota para
esta classificação da operação.
```
### UB27- 10 55/65

```
Se informado grupo de Redução de Alíquota (gIBSUF/gRed):
Percentual de Redução de Alíquota (pRedAliq) não é válido para este
cClassTrib (IBSCBSSel/cClassTrib)
```
```
Obrig. 1034
```
```
Rejeição: Percentual de redução de alíquota da UF não é
válido para este cClassTrib
```
### UB28- 10 55/65

```
Se informado grupo de Redução de Alíquota (gIBSUF/gRed):
Alíquota Efetiva (tag: pAliqEfet) deve ser o resultado da aplicação do
percentual de redução da alíquota (tag: pRedAliq) na alíquota do IBS da
UF (tag: pIBSUF).
Exemplo:
Redução de 40% na alíquota:
Alíquota vigente (A): 10%
Redução na alíquota (R): 40%
Alíquota Efetiva (E): E = A * (1 - R)
E = 10 * (1 - 0,4) = 6
```
```
Obrig. 1035
Rejeição: Valor da Alíquota Efetiva do IBS da UF calculado
incorretamente
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

```
Campo-
Seq
Modelo Regra de Validação Aplic. Msg Descrição Erro
```
### UB29- 10 55/65

```
Não é permitido o uso de desoneração (grupo: gDeson) para este
cClassTrib.
Obrig. 1036
Rejeição: Não é permitido o uso de Desoneração para esta
classificação da operação.
```
```
UB29- 20 55/65 É exigido o uso de desoneração (grupo: gDeson) para este cClassTrib. Obrig. 1037
Rejeição: É exigido o uso de Desoneração da UF para esta
classificação da operação.
```
### UB30- 10 55/65

```
Se informado grupo da desoneração (gIBSUF/gDeson):
O CST for informado, este deve existir na tabela de Classificação
Tributária do IBS/CBS (tag: gDeson/CST)
```
```
Obrig. 1038
Rejeição: CST informado na desoneração da UF
inexistente
```
### UB31- 10 55/65

```
Se informado grupo da desoneração (gIBSUF/gDeson):
O cClassTrib for informado, este deve existir na tabela de Classificação
Tributária do IBS/CBS (tag: gDeson/cClassTrib)
```
```
Obrig. 1039
Rejeição: Classificação Tributária informada na
desoneração da UF inexistente
```
### UB34- 10 55/65

```
Se informado grupo da desoneração (gIBSUF/gDeson):
O valor desonerado (vDeson) deverá ser resultante da Base de Cálculo x
Aliquota (gDeson/vBC x gDeson/pAliq)
```
```
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos
```
```
Obrig. 1040
Rejeição: Valor Desonerado do Município diferente de Base
de Cálculo x Alíquota
```
### UB35- 10 55/65

```
Se informado grupo IBS de competência das Unidades Federadas
(gIBSUF):
```
```
O valor do IBS (vIBSUF) deverá ser resultante da Base de Cálculo x
Alíquota (vBC [tag: gIBSCBS/vBC] x pIBSUF) - vCredPres - vDif -
vDevTrib
```
```
Observação 1: Aceitar uma tolerância de 0,01 a mais ou a menos
```
```
Observação 2: Em caso de preenchimento do grupo de redução (pRed)
a alíquota utilizada deverá ser a tag Alíquota Efetiva (pAliqEfet)
```
```
Observação 3: Conforme cClass escolhido o valor do crédito presumido
(vCredPres) deve ser subtraído do total do IBS.
```
```
Obrig. 1041
Rejeição: NF-e com valor do IBS da UF calculado
incorretamente.
```
### UB37- 10 55/65

```
Alíquota do IBS do município (tag: pIBSMun) deve ser igual a 0,05%
para documento com data de emissão nos anos de 2027 e 2028. Art 343
PL 68/24.
```
```
Obrig. 1042
```
```
Rejeição: Alíquota do IBS da Município deve ser igual a
0,05% para documento emitido em 2027 e 2028
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

```
Campo-
Seq
Modelo Regra de Validação Aplic. Msg Descrição Erro
```
### UB39- 10 55/65

```
Se informado grupo IBS de competência dos Municipios (gIBSMun)
O valor bruto do tributo do IBS dos Municípios deve ser igual à
multiplicação da Base de Cálculo pela Alíquota.
```
```
vTribOP (tag: gIBSMun/vTribOP) = vBC (tag: gIBSCBS/vBC) * pIBSMun
(tag: gIBSMun/pIBSMun)
```
```
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos.
```
```
Obrig. 1043
Rejeição: NF-e com valor bruto do tributo do IBS dos
Municípios calculado incorretamente.
```
### UB40- 10 55/65

```
É exigido o uso de diferimento (grupo: gDif) Municipal para este
cClassTrib.
Obrig. 1044
Rejeição: É exigido o uso de Diferimento Municipal para
esta classificação da operação.
```
### UB42- 10 55/65

```
Se informado grupo do Diferimento (gIBSMun/gDif):
O valor do Diferimento (vDif) deverá ser resultante da Base de Cálculo x
Percentual do Diferimento (vBC x pDif)
```
```
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos
```
```
Obrig. 1045
Rejeição: Valor do Diferimento dos Municípios diferente de
Base de Cálculo x Percentual
```
### UB46- 10 55/65

```
Se informado grupo de Redução de Alíquota (gIBSMun/gRed):
Percentual de Redução de Alíquota (pRedAliq) não é válido para este
cClassTrib (IBSCBSSel/cClassTrib)
```
```
Obrig. 1046
Rejeição: Percentual de redução de alíquota dos
Municípios não é válido para este cClassTrib
```
### UB47- 10 55/65

```
Se informado grupo de Redução de Alíquota (gIBSMun/gRed):
O valor da alíquota efetiva (pAliqEfet) deve ser igual a aplicação da
redução da alíquota do IBS dos Municípios (pRedAliq) na alíquota do
IBS dos Municípios (pIBSMun).
Exemplo:
Redução de 40% na alíquota:
Alíquota vigente (A): 10%
Redução na alíquota (R): 40%
Alíquota Efetiva (E): E = A * (1 - R)
E = 10 * (1 - 0,4) = 6
```
```
Obrig. 1047
Rejeição: Valor da Alíquota Efetiva do IBS dos Municípios
calculado incorretamente
```
### UB48- 10 55/65

```
Código de Classificação Tributária exige informação do grupo de
Desoneração Municipal (gIBSMun/gDeson)
```
```
Obrig. 1048
Rejeição: CST informado exige informação de desoneração
dos Municípios
```
### UB49- 10 55/65

```
Se informado grupo da desoneração (gIBSMun/gDeson):
O CST for informado, este deve existir na tabela de Classificação
Tributária do IBS/CBS (tag: gDeson/CST)
```
```
Obrig. 1049
Rejeição: CST informado na desoneração dos Municípios
inexistente
```
### UB50- 10 55/65

```
Se informado grupo da desoneração (gIBSMun/gDeson):
Se cClassTrib for informado, este deve existir na tabela de Classificação
Tributária do IBS/CBS (tag: gDeson/cClassTrib)
```
```
Obrig. 1050
Rejeição: Classificação Tributária informada na
desoneração dos Municípios inexistente
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

```
Campo-
Seq
Modelo Regra de Validação Aplic. Msg Descrição Erro
```
### UB53- 10 55/65

```
Se informado grupo da desoneração (gIBSMun/gDeson):
O valor desonerado (vDeson) deverá ser resultante da Base de Cálculo x
Aliquota (gDeson/vBC x gDeson/pAliq)
```
```
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos
```
```
Obrig. 1051
Rejeição: Valor Desonerado do Município diferente de Base
de Cálculo x Alíquota
```
### UB54- 10 55/65

```
Se informado grupo IBS de competência dos Municípios (gIBSMun):
```
```
O valor do IBS (vIBSMun) deverá ser resultante da Base de Cálculo x
Alíquota (vBC [tag: gIBSCBS/vBC] x pIBSMun) - vCredPres - vDif -
vDeson - vDevTrib
```
```
Observação 1: Aceitar uma tolerância de 0,01 a mais ou a menos
```
```
Observação 2: Em caso de preenchimento do grupo de redução (pRed)
a alíquota utilizada deverá ser a tag Alíquota Efetiva (pAliqEfet)
```
```
Observação 3: Conforme cClass escolhido o valor do crédito presumido
(vCredPres) deve ser subtraído do total do IBS.
```
```
Obrig. 1052
Rejeição: NF-e com valor do IBS dos Municípios calculado
incorretamente.
```
### UB55- 10 55/65

```
Não é permitido o uso de crédito presumido do IBS (grupo:
gIBSCredPres) para este código de classificação tributária (tag:
IBSCBSSel/cClassTrib).
```
```
Obrig. 1053
Rejeição: Não é permitido o uso de Crédito Presumido para
esta classificação da operação.
```
### UB55- 20 55/65

```
É exigido o uso de crédito presumido do IBS (grupo: gIBSCredPres) para
este código de classificação tributária (tag: IBSCBSSel/cClassTrib).
Obrig. 1054
Rejeição: É exigido o uso de Crédito Presumido do IBS
para esta classificação da operação.
```
### UB58- 10 55/65

```
Se informado grupo do crédito presumido (grupo:
gIBSCBS/gIBSCredPres):
O valor do Crédito Presumido (tag: vCredPres) deverá ser resultante da
Base de Cálculo x Percentual do Crédito Presumido (vBC x pCredPres)
```
```
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos
```
```
Obrig. 1055
Rejeição: Valor do Crédito Presumido do IBS diferente de
Base de Cálculo x Percentual
```
### UB59- 10 55/65

```
Se informado Valor do Crédito Presumido em condição suspensiva
(tag:gIBSCredPres/vCredPresCondSus) , validar se código de
classificação do crédito presumido (tag: gIBSCredPres/cCredPres) = 4 -
Aquisição de bens móveis de PF não contrib. para revenda (veículos /
brecho).
```
```
Obrig. 1056
```
```
Rejeição: Valor do Crédito Presumido em condição
suspensiva deve ser informado somente para cCredPres =
4 - Aquisição de bens móveis de PF não contrib. para
revenda (veículos / brecho).
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

```
Campo-
Seq
Modelo Regra de Validação Aplic. Msg Descrição Erro
```
### UB63- 10 55/65

```
Se informado grupo CBS (gCBS)
O valor bruto do tributo da CBS deve ser igual à multiplicação da Base
de Cálculo pela Alíquota.
```
```
vTribOP (tag: gCBS/vTribOP) = vBC (tag: gIBSCBS/vBC) * pCBS (tag:
gCBS/pCBS)
```
```
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos.
```
```
Obrig. 1057
Rejeição: NF-e com valor bruto do tributo da CBS calculado
incorretamente.
```
### UB64- 10 55/65

```
É exigido o uso de crédito presumido da CBS (grupo: gCBSCredPres)
para este cClassTrib.
Obrig. 1058
Rejeição: É exigido o uso de Crédito Presumido da CBS
para esta classificação da operação.
```
### UB67- 10 55/65

```
Se informado grupo do crédito presumido (gCBS/gCBSCredPres):
O valor do Crédito Presumido (vCredPres) deverá ser resultante da Base
de Cálculo x Percentual do Crédito Presumido (vBC x pCredPres)
```
```
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos
```
```
Obrig. 1059
Rejeição: Valor do Crédito Presumido da CBS diferente de
Base de Cálculo x Percentual
```
### UB68- 10 55/65

```
Se informado Valor do Crédito Presumido em condição suspensiva
(tag:gCBSCredPres/vCredPresCondSus) , validar se código de
classificação do crédito presumido (tag: gCBSCredPres/cCredPres) = 4 -
Aquisição de bens móveis de PF não contrib. para revenda (veículos /
brecho).
```
```
Obrig. 1060
```
```
Rejeição: Valor do Crédito Presumido em condição
suspensiva deve ser informado somente para cCredPres =
4 - Aquisição de bens móveis de PF não contrib. para
revenda (veículos / brecho).
```
### UB69- 10 55/65

```
É exigido o uso de diferimento (grupo: gDif) da CBS para este
cClassTrib.
Obrig. 1061
Rejeição: É exigido o uso de Diferimento da CBS para esta
classificação da operação.
```
### UB71- 10 55/65

```
Se informado grupo do Diferimento (gCBS/gDif):
O valor do Diferimento (vDif) deverá ser resultante da Base de Cálculo x
Percentual do Diferimento (vBC x pDif)
```
```
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos
```
```
Obrig. 1062
```
```
Rejeição: Valor do Diferimento da CBS diferente de Base
de Cálculo x Percentual
```
### UB75- 10 55/65

```
Se informado grupo de Redução de Alíquota (gCBS/gRed):
Percentual de Redução de Alíquota (pRedAliq) não é válido para este
cClassTrib (IBSCBSSel/cClassTrib)
```
```
Obrig. 1063
```
```
Rejeição: Percentual de redução de alíquota da CBS não é
válido para este cClassTrib
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

```
Campo-
Seq
Modelo Regra de Validação Aplic. Msg Descrição Erro
```
### UB76- 10 55/65

```
Se informado grupo de Redução de Alíquota (gCBS/gRed):
O valor da alíquota efetiva (pAliqEfet) deve ser igual a aplicação da
redução da alíquota da CBS (pRedAliq) na alíquota da CBS (pCBS).
Exemplo:
Redução de 40% na alíquota:
Alíquota vigente (A): 10%
Redução na alíquota (R): 40%
Alíquota Efetiva (E): E = A * (1 - R)
E = 10 * (1 - 0,4) = 6
```
```
Obrig. 1064
Rejeição: Valor da Alíquota Efetiva da CBS calculado
incorretamente
```
### UB77- 10 55/65

```
Código de Classificação Tributária exige informação do grupo de
Desoneração da CBS (gCBS/gDeson)
Obrig. 1065
Rejeição: CST informado exige informação de desoneração
da CBS
```
### UB78- 10 55/65

```
Se informado grupo da desoneração (gCBS/gDeson):
Se o CST for informado, este deve existir na tabela de Classificação
Tributária do IBS/CBS (tag: gDeson/CST)
```
```
Obrig. 1066
Rejeição: CST informado na desoneração da CBS
inexistente
```
### UB79- 10 55/65

```
Se informado grupo da desoneração (gCBS/gDeson):
Se o cClassTrib for informado, este deve existir na tabela de
Classificação Tributária do IBS/CBS (tag: gDeson/cClassTrib)
```
```
Obrig. 1067
Rejeição: Classificação Tributária informada na
desoneração a CBS inexistente
```
### UB82- 10 55/65

```
Se informado grupo da desoneração (gCBS/gDeson):
O valor desonerado (vDeson) deverá ser resultante da Base de Cálculo x
Aliquota (gDeson/vBC x gDeson/pAliq)
```
```
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos
```
```
Obrig. 1068
Rejeição: Valor Desonerado da CBS diferente de Base de
Cálculo x Alíquota
```
### UB83- 10 55/65

```
Se informado grupo CBS (gCBS):
```
```
O valor da CBS (vCBS) deverá ser resultante da Base de Cálculo x
Alíquota (vBC [tag: gIBSCBS/vBC] x pCBS) - vCredPres - vDif - vDevTrib
```
```
Observação 1: Aceitar uma tolerância de 0,01 a mais ou a menos
```
```
Observação 2: Em caso de preenchimento do grupo de redução (pRed)
a alíquota utilizada deverá ser a tag Alíquota Efetiva (pAliqEfet)
```
```
Observação 3: Conforme cClass escolhido o valor do crédito presumido
(vCredPres) deve ser subtraído do total do IBS.
```
```
Obrig. 1069
Rejeição: NF-e com valor da CBS calculado
incorretamente.
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

```
Campo-
Seq
Modelo Regra de Validação Aplic. Msg Descrição Erro
```
### UB104- 10 55/65

```
Se informado grupo do IBS e CBS monofásico (grupo: gIBSCBSMono):
O valor total do IBS Monofásico (tag: vTotIBSMono) deverá ser
resultante de:
vIBSMono + vIBSMonoReten -vIBSMonoDif
```
```
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos
```
```
Obrig. 1070
```
```
Rejeição: Valor do IBS monofásico calculado
incorretamente.
```
### UB105- 10 55/65

```
Se informado grupo do IBS e CBS monofásico (grupo: gIBSCBSMono):
O valor total da CBS Monofásica (tag: vTotCBSMono) deverá ser
resultante de:
vCBSMono + vCBSMonoReten -vCBSMonoDif
```
```
Observação: Aceitar uma tolerância de 0,01 a mais ou a menos
```
```
Obrig. 1071
Rejeição: Valor da CBS monofásico calculado
incorretamente.
```
9.4. Grupo VB. Referenciamento de item de outro Documento Fiscal Eletrônico - DF-e

```
Campo-
Seq
Modelo Regra de Validação Aplic. Msg Descrição Erro
```
### VB02- 10 55/65

```
Item referenciado em duplicidade.
Informado mais de uma vez uma mesma chave de acesso (tag:
DFeReferenciado/chaveAcesso) e item (tag: DFeReferenciado/nItem).
```
```
Obrig. 1072 Rejeição: Item referenciado em duplicidade.
```
### VB02- 20 55/65

```
Um único documento fiscal deve ser referenciado.
Informado chaves de acesso diferentes em documento referenciado (tag:
DFeReferenciado/chaveAcesso).
```
```
Obrig. 1072 Rejeição: Mais de um documento fiscal referenciado.
```
9.5. Grupo W03. Total da NF-e - IBS / CBS / IS

```
Campo-
Seq
Modelo Regra de Validação Aplic. Msg Descrição Erro
```
### W31- 10 55/65

```
O grupo de totais do IBS, CBS e IS (IBSCBSSelTot) só deve ser
informado se existir pelo menos uma ocorrência de IBS, CBS ou Imposto
Seletivo nos itens.
```
```
Obrig. 1073
```
```
Rejeição: Total de IBS, CBS e IS só deve ser informado se
existir IBS/CBS declarado nos itens do DFe
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

```
Campo-
Seq
Modelo Regra de Validação Aplic. Msg Descrição Erro
```
### W33- 10 55/65

```
O total da BC do imposto seletivo deverá ser a soma dos campos vBC
(tag: seletivo/vBC) informados nos itens.
```
```
Obrig. 1074
Rejeição: Total da BC do Imposto Seletivo difere da soma
dos itens
```
```
W34- 10 55/65
O total do Imposto Seletivo deverá ser a soma dos campos vImpSel (tag:
seletivo/vImpSel) informados nos itens.
Obrig. 1075
Rejeição: Total do Imposto Seletivo difere da soma dos
itens
```
```
W35- 10 55/65
O total da BC do IBS e da CBS de deverá ser a soma dos campos vBC
(tag: gIBSCBS/vBC) informados nos itens
Obrig. 1076
Rejeição: Total da BC do IBS e da CBS difere da soma dos
itens
```
```
W38- 10 55/65
O total do diferimento do IBS UF deverá ser a soma do campo vDif do
IBS UF informados nos itens
```
```
Obrig. 1077
Rejeição: Total de Diferimento do IBS UF difere da soma
dos itens
```
```
W39- 10 55/65
```
```
O total devolvido do IBS UF deverá ser a soma do campo vDevTrib do
IBS UF informados nos itens
Obrig. 1078
```
```
Rejeição: Total Devolvido do IBS UF difere da soma dos
itens
```
```
W40- 10 55/65
O total desonerado do IBS UF deverá ser a soma do campo vDeson do
IBS UF informados nos itens
Obrig. 1079
Rejeição: Total Desonerado do IBS UF difere da soma dos
itens
```
```
W41- 10 55/65
O total do IBS UF deverá ser a soma do campo vIBSUF informados nos
itens
```
```
Obrig. 1080 Rejeição: Total de IBS UF difere da soma dos itens
```
### W43- 10 55/65

```
O total do Diferimento do IBS Municipal deverá ser a soma do campo
vDif do IBS Municipal informados nos itens
Obrig. 1081
```
```
Rejeição: Total de Diferimento do IBS Municipal difere da
soma dos itens
```
```
W44- 10 55/65
O total devolvido do IBS Municipal deverá ser a soma do campo
vDevTrib do IBS Municipal informados nos itens
Obrig. 1082
Rejeição: Total Devolvido do IBS Municipal difere da soma
dos itens
```
```
W45- 10 55/65
O total desonerado do IBS Municipal deverá ser a soma do campo
vDeson do IBS Municipal informados nos itens
Obrig. 1083
Rejeição: Total Desonerado do IBS Municipal difere da
soma dos itens
```
```
W46- 10 55/65
O total do IBS Municipal deverá ser a soma do campo vIBSMun
informados nos itens
```
```
Obrig. 1084 Rejeição: Total de IBS Municipal difere da soma dos itens
```
### W47- 10 55/65

```
O total do IBS deverá ser a soma do IBS das UF e do IBS Municipal
(vIBSUF + vIBSMun)
Obrig. 1085
Rejeição: Total do IBS difere da soma do IBS UF e IBS
Municipal
```
```
W48- 10 55/65
O total do Crédito Presumido do IBS deverá ser a soma do campo
vCredPres do IBS informados nos itens
Obrig. 1086
Rejeição: Total de Crédito Presumido do IBS UF difere da
soma dos itens
```
```
W51- 10 55/65
O total do Crédito Presumido do CBS deverá ser a soma do campo
vCredPres do CBS informados nos itens
```
```
Obrig. 1087
Rejeição: Total de Crédito Presumido do CBS difere da
soma dos itens
```
```
W53- 10 55/65
O total do Diferimento do CBS deverá ser a soma do campo vDif do CBS
informados nos itens
Obrig. 1088
Rejeição: Total de Diferimento do CBS difere da soma dos
itens
```
```
W54- 10 55/65
O total devolvido do CBS deverá ser a soma do campo vDevTrib do CBS
informados nos itens
Obrig. 1089 Rejeição: Total Devolvido do CBS difere da soma dos itens
```
### W55- 10 55/65

```
O total desonerado do CBS deverá ser a soma do campo vDeson do
CBS informados nos itens
```
```
Obrig. 1090
Rejeição: Total Desonerado do CBS difere da soma dos
itens
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

```
Campo-
Seq
Modelo Regra de Validação Aplic. Msg Descrição Erro
```
```
W56- 10 55/65 O total do CBS deverá ser a soma do campo vCBS informados nos itens Obrig. 1091 Rejeição: Total de CBS difere da soma dos itens
```
### W58- 10 55/65

```
O total do IBS monofásico deverá ser a soma dos campos
vTOTIBSMono informados nos itens
Obrig. 1092 Rejeição: Total do IBS monofásico difere da soma dos itens
```
### W59- 10 55/65

```
O total da CBS monofásica deverá ser a soma dos campos
vTOTCBSMono informados nos itens
Obrig. 1093
Rejeição: Total da CBS monofásica difere da soma dos
itens
```
### W60- 10 55/65

```
O total geral do DFe deverá ser a soma do total NF (grupo total/vNF) +
vTotIBS (IBSCBSSelTot/gIBS/vTotIBS) + vCBS
(IBSCBSSelTot/gCBS/vCBS)+ vSel (IBSCBSSelTot/gSel/vImpSel) +
vTotIBSMono (gIBSCBSMono/vTotIBSMono) +
(gIBSCBSMono/vTotCBSMono) vTotCBSMono
```
```
Obrig. 1094
```
```
Rejeição: Total da NFe difere da soma do total da Nota
Fiscal, IBS, CBS e IS
```
3A. Banco de Dados: NF-e Referenciada

```
Campo-Seq Modelo Regra de Validação Aplic. Msg Descrição Erro
```
### 3BA02- 10 55

```
Para cada NF-e referenciada (tag:refNFe), se a UF da Chave de Acesso
referenciada for igual a UF do Emitente:
```
- Acessar BD NFE com Chave de Acesso referenciada (se mod=55)
- NF-e referenciada inexistente

```
Exceção: A NF-e referenciada pode não existir no caso de Emissão em
Contingência (tpEmis = 2, 4 ou 5) (NT 2013/003), desde que a Chave de
Acesso da NF-e referenciada tenha o Ano-Mês de Emissão inferior a 1
mês da data atual ou desde que exista o EPEC (NT 2021.004). Esta
exceção não se aplica para NF-e com finalidade de crédito (tag: finNF-e =
5).
Observação 1 : A exceção acima não se aplica para “finNFe=2" (NF-e
Complementar).
Observação 2: Regra de validação se aplica obrigatoriamente para
“finNFe=5" (NF-e de cŕedito).
```
```
Facult 267
```
```
Rejeição: Chave de Acesso referenciada inexistente [nRef:
xxx]
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

```
Campo-Seq Modelo Regra de Validação Aplic. Msg Descrição Erro
```
### 3BA02- 70 55

```
Se NF-e de crédito (tag: finNFe = 5) com indicador de juros / multa (tag:
indMultaJuros):
Chave de acesso referenciada deve estar autorizada ter finalidade de
débito (tag: finNFe = 6) e ter indicador de multa/ juros (tag:
indMultaJuros).
Nota Explicativa: Em operação entre empresas, o fornecedor pode
cobrar multa e juros do cliente e no valor cobrado de juros e multas,
incide o IBS e a CBS.
O cliente ao pagar este valor adicional pode se creditar do imposto pago.
Nesta situação pode ser emitida uma nota com o acerto dos juros e
multas.
A cobrança de juros e multas podem ser provenientes de atraso de
pagamento no boleto bancário, sem que haja uma nota para juros e
multas.
Nessa situação deve haver um evento para que o adquirente possa
garantir o crédito adicional, gerando um débito para o outro contribuinte.
```
```
Obrig. 1095
```
```
Rejeição: Nota de crédito de multa/juros deve ter NF-e de
débito de multa/juros referenciada.
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

10. Eventos

10.1. Lista de eventos

Esta NT cria os eventos a seguir para a apuração do IBS e da CBS.

CÓDIGO EVENTO

```
CÓDIGO
CANC.
```
```
EVENTO DE
CANCELAMENTO
```
Autor

112110

```
Informação de efetivo pagamento integral para liberar crédito
presumido do adquirente
```
112111

```
Cancelamento do evento
112110
```
Emitente

211110 Solicitação de Apropriação de crédito presumido 211111

```
Cancelamento do evento
211110
```
Destinatário

211120 Destinação de item para consumo pessoal 211121

```
Cancelamento do evento
211120
```
Destinatário

211130 Imobilização de Item 211131

```
Cancelamento do evento
211130
```
Destinatário

211140 Solicitação de Apropriação de Crédito de Combustível 211141

```
Cancelamento do evento
211140
```
Destinatário

412110

```
Solicitação de Apropriação de Crédito de fornecimento de bem
móvel usado
```
412111

```
Cancelamento do evento
412110
```
Fisco

211150

```
Solicitação de Apropriação de Crédito para bens e serviços que
dependem de atividade do adquirente
```
211151

```
Cancelamento do evento
211150
```
Destinatário

212110

```
Manifestação sobre Pedido de Transferência de Crédito de IBS
em Operações de Sucessão
```
212111

```
Cancelamento do evento
212110
```
Sucessora

212120

```
Manifestação sobre Pedido de Transferência de Crédito CBS em
Operações de Sucessão
```
212121

```
Cancelamento do evento
212120
```
Sucessora

412120

```
Manifestação do Fisco sobre Pedido de Transferência de Crédito
de IBS em Operações de Sucessão
```
412121

```
Cancelamento do evento
412120
```
Fisco

412130

```
Manifestação do Fisco sobre Pedido de Transferência de Crédito
de CBS em Operações de Sucessão
```
412131

```
Cancelamento do evento
412130
```
Fisco


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

10.2. Registro de Eventos

O Web Service de Registro de Evento possui uma parte geral, complementada por uma área específica para cada tipo de evento.

A parte geral se mantém a mesma para envio de todos os eventos e está especificada na seção 5.8 do Manual de Orientação do Contribuinte e na

seção Web Service – NFeRecepcaoEvento – Parte Geral do MOC Online.

A mensagem de retorno será modificada, conforme apresentado na seção 7.2.1 e o leiaute das partes específicas dos novos eventos está especificado

nas seções seguintes.

10.2.1. Leiaute Mensagem de Retorno da Parte Geral

O Web Service de Registro de Evento possui uma interface genérica complementada por uma área específica para cada tipo de evento. Segue o

leiaute da mensagem de retorno (resposta)

Schema XML: retEnvEventoNFe_v1.0.xsd

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição / Observações
R01 envEvento Raiz - - - - TAG raiz
R02 versão A R01 N 1 - 1 2v2 Versão do leiaute
R03 idLote E R01 N 1 - 1 1 - 15 Idem a mensagem de entrada
R04 tpAmb E R01 N 1 - 1 1 Idem a mensagem de entrada
R05 verAplic E R01 C 1 - 1 1 - 20 Versão da aplicação que processou o evento.
R06 cOrgao E R01 N 1 - 1 2 Órgão da recepção do Evento, idem a mensagem de entrada.
```
```
R07 cStat E R01 N 1 - 1 3
Código do status da resposta para o Lote de Eventos. Se não tiver erro, será
retornado: “128-Lote de Evento Processado“.
R08 xMotivo E R01 C 1 - 1 1 - 255 Descrição do status da resposta
R09 retEvento G R01 - 0 - 20 - Grupo do resultado do processamento do para cada Evento
R10 versão A R09 N 1 - 1 2v2 Versão do leiaute
R11 infEvento G R09 - 1 - 1 - Grupo de informações do registro do Evento.
```
```
R12 id ID R11 C 0 - 1 17
```
```
Identificador da TAG a ser assinada, somente deve ser informado se o órgão de
registro assinar a resposta. No caso de assinatura, preencher com o número do
protocolo, precedido pela literal “ID”.
R13 tpAmb E R11 C 1 - 1 1 Idem a mensagem de entrada
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição / Observações
```
```
R14 verAplic E R11 N 1 - 1 1 - 20
Versão da aplicação que registrou o Evento, utilizar literal que permita a
identificação do órgão, como a sigla da UF ou do órgão.
R15 cOrgao E R11 N 1 - 1 2 Idem a mensagem de entrada
R16 cStat E R11 N 1 - 1 4 Código do status da resposta.
R17 xMotivo E R11 C 1 - 1 1 - 255 Descrição do status da resposta
R18 chNFe E R11 N 1 - 1 44 Idem a mensagem de entrada
R19 tpEvento E R11 N 0 - 1 6 Idem a mensagem de entrada
R20 xEvento E R11 C 0 - 1 5 - 60 Idem a mensagem de entrada
R21 nSeqEvento E R11 N 0 - 1 1 - 2 Idem a mensagem de entrada
```
```
R50 dhRegEvento E R11 D 1 - 1 -
```
```
Data e hora de registro do evento no formato AAAA-MM-DDTHH:MM:SS TZD
(formato UTC). Se o evento for rejeitado informar a data e hora de recebimento do
evento.
```
```
R51 nProt E R11 N 0 - 1 15
```
```
Número Protocolo do Evento 1 posição (1- Secretaria da Fazenda Estadual, 2-
RFB, 3-SVRS), 2 posições para o código da UF, 2 posições para o ano e 10
posições para o sequencial no ano.
```
```
P91 Signature G R09 XML 1 - 1 1
Assinatura Digital do documento XML, a assinatura deverá ser aplicada no
elemento infEvento.
```
10.2.2. Evento: Informação de efetivo pagamento integral para liberar crédito presumido do adquirente

Função: Permitir que o emitente da NFe informe o efetivo pagamento integral a fim de liberar crédito presumido do adquirente

Modelo: NF-e modelo 55

Autor do Evento: Emitente da NFe (emitente)

Código do Tipo de Evento: 112110

10.2.2.1. Leiaute Mensagem de Entrada

Entrada: Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos

especificada na seção Erro! Fonte de referência não encontrada. do MOC.


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

Schema XML: envEventoNFe_v9.99.xsd

Schema XML - parte específica: e112110_v1.00.xsd

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 - Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 85 Descrição do evento: "Informação de efetivo pagamento integral para liberar crédito presumido do adquirente"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código do Órgão Autor do Evento. Informar o Código da UF para este Evento.
```
```
P21 tpAutor E P17 N 1 - 1 1
Informar 1=Empresa emitente.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco; 6=RFB; 9=Outros Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 indQuitacao E P17 N 1 - 1 1 Indicador de efetiva quitação do pagamento integral referente a NFe referenciada. Valor deve ser igual a "1"
```
10.2.2.2. Leiaute Mensagem de Retorno

Retorno: Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral,

especificado no item 7.2.1.

10.2.3. Evento: Solicitação de Apropriação de crédito presumido

Função: Evento a ser gerado pelo adquirente em relação às notas fiscais de aquisição de emissão de terceiros e que lhe gerem o direito à apropriação

de crédito presumido.

Autor: Adquirente/Destinatário (quando os dois estiverem preenchidos, devem ser iguais) da nota fiscal

Modelo: NF-e modelo 55

Código do Tipo de Evento: 211110

10.2.3.1. Leiaute Mensagem de Entrada

Entrada: Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos

especificada na seção Erro! Fonte de referência não encontrada. do MOC.


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

Schema XML: envEventoNFe_v9.99.xsd

Schema XML - parte específica: e 211110 _v1.00.xsd

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 - Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 47 Descrição do evento: "Solicitação de Apropriação de crédito presumido"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código do Órgão Autor do Evento. Informar o Código da UF para este Evento.
```
```
P21 tpAutor E P17 N 1 - 1 1
```
```
Informar 1=Empresa emitente.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco;
6=RFB; 9=Outros Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 gCredPres G P17 N 1 - N - Informações de crédito presumido por item
P24 nitem A P23 N 1 - 1 1 - 3 Corresponde ao atributo “nItem” do elemento “det” do documento referenciado.
P25 vBC A P23 N 1 - 1 1 Valor do base de cálculo do item
P26 gIBS G P23 0 - 1 - Grupo de Informações do Crédito Presumido do IBS
```
```
P27 cCredPres E P27 N 1 - 1 2
Código de Classificação do Crédito presumido, conforme tabela de
CÓDIGO DE CLASSIFICAÇÃO DO CRÉDITO PRESUMIDO
P28 pCredPres E P27 N 1 - 1 3v2- 4 Percentual do Crédito Presumido
P29 vCredPres E P27 N 1 - 1 13v2 Valor do Crédito Presumido
```
10.2.3.2. Leiaute Mensagem de Retorno

Retorno: Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral,

especificado no item 7.2.1.

10.2.3.3. Validação das Regras de Negócio **–** Específicas

Serão aplicadas as regras de validação gerais apresentadas no item Erro! Fonte de referência não encontrada. do MOC e as regras de negócio

específicas listadas a seguir.


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

```
# Regra de Validação Aplic. Msg Descrição Erro
Banco de Dados: NF-e
```
### 2P21- 10

```
Se tpAutor=2-Empresa Destinatário:
```
- CNPJ/CPF do Autor diverge do CNPJ/CPF do
Destinatário da NF-e

```
Obrig. 575 Rejeição: Autor do evento diverge do destinatário da NF-e
```
### 2P24- 10

```
Acessar BD e verificar se número do item do evento (tag:
gCredPres/nItem) informado existe na NFe referenciada (tag:
det/nItem)
```
```
Obrig. 1096 Rejeição: Número de item não existe na NFe
```
10.2.4. Evento: Destinação de item para consumo pessoal

Função: Permitir ao adquirente informar quando uma aquisição for destinada para o consumo de pessoa física, hipótese em que não haverá direito à

apropriação de crédito.

Modelo: NF-e modelo 55

Autor do Evento: Destinatário da NF-e. A mensagem XML do evento será assinada com o certificado digital que tenha o CNPJ-Base (8 primeiras

posições do CNPJ) ou CPF do Destinatário da NF-e.

Código do Tipo de Evento: 211120

10.2.4.1. Leiaute Mensagem de Entrada

Entrada: Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos

especificada na seção Erro! Fonte de referência não encontrada. do MOC.


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

Schema XML: envEventoNFe_v9.99.xsd

Schema XML - parte específica: e 211120 _v1.00.xsd

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 - Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 39 Descrição do evento: "Destinação de item para consumo pessoal"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código do Órgão Autor do Evento. Informar o Código da UF para este Evento.
```
```
P21 tpAutor E P17 N 1 - 1 1
```
```
Informar 2=Empresa destinatária.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco;
6=RFB; 9=Outros Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 gConsumo G P17 N 1 - N 1 Informações de crédito a ser estornado por item
P24 nItem A P23 N 1 - 1 1 Corresponde ao atributo “nItem” do elemento “det” do documento referenciado.
P25 vIBS A P23 N 1 - 1 13v2 Valor do IBS a ser estornado
P26 vCBS A P23 N 1 - 1 13v2 Valor da CBS a ser estornado
P27 gControleEstoque G P23 N 1 - 1 1 Informações de crédito presumido por item
P28 qConsumo E P27 N 1 - 1 11v0- 4 Informar a quantidade para consumo de pessoa física
P29 uConsumo E P27 C 1 - 1 1 - 6 Informar a unidade relativa ao campo gConsumo
```
10.2.4.2. Leiaute Mensagem de Retorno

Retorno: Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral,

especificado no item 7.2.1.


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

10.2.4.3. Validação das Regras de Negócio **–** Específicas

Serão aplicadas as regras de validação gerais apresentadas no item Erro! Fonte de referência não encontrada. do MOC e as regras de negócio

específicas listadas a seguir.

```
# Regra de Validação Aplic. Msg Descrição Erro
```
### 2P21- 10

```
Se tpAutor=2-Empresa Destinatário:
```
- CNPJ/CPF do Autor diverge do CNPJ/CPF do
Destinatário da NF-e

```
Obrig. 575 Rejeição: Autor do evento diverge do destinatário da NF-e
```
### 2P24- 10

```
Acessar BD e verificar se número do item do evento (tag:
gconsumo/nItem) informado existe na NFe referenciada (tag:
det/nItem)
```
```
Obrig. 1096 Rejeição: Número de item não existe na NFe
```
### 2P25- 10

```
Acessar BD e verificar se valor do IBS do evento (tag:
gConsumo/vIBS) é maior que o valor do IBS do item
informado na NFe
```
```
Obrig. 1097
```
```
Rejeição: O valor do IBS do item não pode ser maior que o valor do IBS
do respectivo item na NFe.
```
### 2P26- 10

```
Acessar BD e verificar se valor da CBS do evento (tag:
gConsumo/vCBS) é maior que o valor da CBS do item
informado na NFe
```
```
Obrig. 1098
Rejeição: O valor da CBS do item não pode ser maior que o valor da CBS
do respectivo item na NFe.
```
### 2P28- 10

```
Se unidade de consumo for igual a unidade da nota, acessar
BD e verificar se a quantidade de consumo do item informado
(tag: gControleEstoque/qConsumo) é maior que a quantidade
no item da NFe (tag: det/prod/qCom)
```
```
Obrig. 1099
Rejeição: A quantidade de consumo não pode ser maior que a quantidade
do respectivo item na NFe (qCom)
```
10.2.5. Evento: Imobilização de Item

Função: Evento a ser gerado pelo adquirente de bem, quando este for integrado ao seu ativo imobilizado, a fim de viabilizar a adequada identificação,

pelos sistemas da administração tributária, de prazo-limite para apreciação de eventuais pedidos de ressarcimento do respectivo crédito, nos termos

do art. 59, I do PLP 68/2024.

Modelo: NF-e modelo 55

Autor do Evento: Destinatário da NF-e (Adquirente). A mensagem XML do evento será assinada com o certificado digital que tenha o CNPJ-Base (8

primeiras posições do CNPJ) ou CPF do Destinatário da NF-e.

Código do Tipo de Evento: 211130


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

10.2.5.1. Leiaute Mensagem de Entrada

Entrada: Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos

especificada na seção Erro! Fonte de referência não encontrada. do MOC.

Schema XML: envEventoNFe_v9.99.xsd

Schema XML - parte específica: e211130_v1.00.xsd

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 - Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 5 - 60 Descrição do evento: "Imobilização de Item"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código da UF do emitente do Evento
```
```
P21 tpAutor E P17 N 1 - 1 1
```
```
Informar 2=Empresa destinatária.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco;
6=RFB; 9=Outros Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 gImobilizacao G P17 N 1 - N 1 Informações de itens integrados ao ativo imobilizado
P24 nitem A P23 N 1 - 1 1 Corresponde ao atributo “nItem” do elemento “det” do documento referenciado.
P25 vIBS A P23 N 1 - 1 13v2 Valor do IBS relativo a imobilização
P26 vCBS A P23 N 1 - 1 13v2 Valor da CBS relativo a imobilização
P26 gControleEstoque G P26 N 1 - 1 1 Informações de crédito presumido por item
P27 qImobilizada E P26 N 1 - 1 11v0- 4 Informar a quantidade do item a ser imabilizado
P28 uImobilizada E P26 C 1 - 1 1 - 6 Informar a unidade relativa ao campo qImobilizada
```
10.2.5.2. Leiaute Mensagem de Retorno

Retorno: Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral,

especificado no item 7.2.1.


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

10.2.5.3. Validação das Regras de Negócio **–** Específicas

Serão aplicadas as regras de validação gerais apresentadas no item Erro! Fonte de referência não encontrada. do MOC e as regras de negócio

específicas listadas a seguir.

```
# Regra de Validação Aplic. Msg Descrição Erro
Banco de Dados: NF-e
```
### 2P21- 10

```
Se tpAutor=2-Empresa Destinatário:
```
- CNPJ/CPF do Autor diverge do CNPJ/CPF do
Destinatário da NF-e

```
Obrig. 575 Rejeição: Autor do evento diverge do destinatário da NF-e
```
### 2P24- 10

```
Acessar BD e verificar se número do item do evento (tag:
gImobilizacao/nItem) informado existe na NFe referenciada
(tag: det/nItem)
```
```
Obrig. 1096 Rejeição: Número de item não existe na NFe
```
### 2P25- 10

```
Acessar BD e verificar se valor do IBS do evento (tag:
gImobilizacao/vIBS) é maior que o valor do IBS do item
informado na NFe
```
```
Obrig. 1097
Rejeição: O valor do IBS do item não pode ser maior que o valor do IBS
do respectivo item na NFe.
```
### 2P26- 10

```
Acessar BD e verificar se valor da CBS do evento (tag:
gImobilizacao /vCBS) é maior que o valor da CBS do item
informado na NFe
```
```
Obrig. 1098
Rejeição: O valor da CBS do item não pode ser maior que o valor da CBS
do respectivo item na NFe.
```
### 2P28- 10

```
Se unidade for igual a unidade da nota, acessar BD e verificar
se a quantidade de consumo do item informado (tag:
gControleEstoque/qImobilizada) é maior que a quantidade no
item da NFe (tag: det/prod/qCom)
```
```
Obrig. 1100
Rejeição: A quantidade de consumo não pode ser maior que a quantidade
do respectivo item na NFe (qCom)
```
10.2.6. Evento: Solicitação de Apropriação de Crédito de Combustível

Função: Evento a ser gerado pelo adquirente de combustível listado no art. 167 do PLP 68/2024 e que pertença à cadeia produtiva desses

combustíveis, para solicitar a apropriação de crédito referente à parcela que for consumida em sua atividade comercial.

Modelo: NF-e modelo 55

Autor do Evento: Destinatário da NF-e (Adquirente de combustível que faça parte da cadeia produtiva de combustíveis). A mensagem XML do evento

será assinada com o certificado digital que tenha o CNPJ-Base (8 primeiras posições do CNPJ) ou CPF do Destinatário da NF-e.

Código do Tipo de Evento: 211140

10.2.6.1. Leiaute Mensagem de Entrada


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

Entrada: Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos

especificada na seção Erro! Fonte de referência não encontrada. do MOC.

Schema XML: envEventoNFe_v9.99.xsd

Schema XML - parte específica: e211140_v1.00.xsd

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 - Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 52 Descrição do evento: "Solicitação de Apropriação de Crédito de Combustível"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código da UF do emitente do Evento
```
```
P21 tpAutor E P17 N 1 - 1 1
```
```
Informar 2=Empresa destinatária.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco;
6=RFB; 9=Outros Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 gConsumoComb G P17 N 1 - N 1 Informações de consumo de combustíveis
P24 nitem A P23 N 1 - 1 1 Corresponde ao atributo “nItem” do elemento “det” do documento referenciado.
P25 vIBS A P23 N 1 - 1 13v2 Valor do IBS relativo ao consumo de combustível
P26 vCBS A P23 N 1 - 1 13v2 Valor da CBS relativo ao consumo de combustível
P27 gControleEstoque G P26 N 1 - 1 1 Informações de crédito presumido por item
P28 qComb A P27 N 1 - 1 11v0- 4 Informar a quantidade de consumo do item
P28 uComb G P27 N 1 - N 1 Informar a unidade relativa ao campo qComb
```
10.2.6.2. Leiaute Mensagem de Retorno

Retorno: Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral,

especificado no item 7.2.1.

10.2.6.3. Validação das Regras de Negócio **–** Específicas

Serão aplicadas as regras de validação gerais apresentadas no item Erro! Fonte de referência não encontrada. do MOC e as regras de negócio

específicas listadas a seguir.


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

```
# Regra de Validação Aplic. Msg Descrição Erro
Banco de Dados: NF-e
```
```
2P21- 10
```
```
Se tpAutor=2-Empresa Destinatário:
```
- CNPJ/CPF do Autor diverge do CNPJ/CPF do Destinatário da NF-e
    Obrig. 575

```
Rejeição: Autor do evento diverge do destinatário da
NF-e
```
### 2P24- 10

```
Acessar BD e verificar se número do item do evento (tag:
gConsumoComb/nItem) informado existe na NFe referenciada (tag:
det/nItem)
```
```
Obrig. 1096 Rejeição: Número de item não existe na NFe
```
### 2P25- 10

```
Acessar BD e verificar se valor do IBS do evento (tag:
gConsumoComb/vIBS) é maior que o valor do IBS do item informado na
NFe
```
```
Obrig. 1097
Rejeição: O valor do IBS do item não pode ser maior
que o valor do IBS do respectivo item na NFe.
```
### 2P26- 10

```
Acessar BD e verificar se valor da CBSdo evento (tag:
gConsumoComb/vCBS) é maior que o valor da CBS do item informado na
NFe
```
```
Obrig. 1098
Rejeição: O valor da CBS do item não pode ser maior
que o valor da CBS do respectivo item na NFe.
```
### 2P28- 10

```
Se unidade de consumo for igual a unidade da nota, acessar BD e verificar
se a quantidade de consumo do item informado (tag:
gConsumoComb/qComb) é maior que a quantidade no item da NFe (tag:
det/prod/qCom)
```
```
Obrig. 1101
```
```
Rejeição: A quantidade do item a ser imobilizado não
pode ser maior que a quantidade do respectivo item na
NFe (qCom)
```
10.2.7. Evento: Solicitação de Apropriação de Crédito de fornecimento de bem móvel usado

Função: Evento a ser gerado pelo fisco sempre que for emitida nota fiscal de venda de bem móvel usado por contribuinte do regime regular de

apuração, que referencie nota fiscal de aquisição em que tenha havido solicitação de apropriação de crédito presumido sob condição resolutória.

Modelo: NF-e modelo 55

Autor do Evento: Fisco (evento de propagação)

Código do Tipo de Evento: 412110

10.2.7.1. Leiaute Mensagem de Entrada

Entrada: Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos

especificada na seção Erro! Fonte de referência não encontrada. do MOC.


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

Schema XML: envEventoNFe_v9.99.xsd

Schema XML - parte específica: e412110_v1.00.xsd

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 - Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 72 Descrição do evento: "Solicitação de Apropriação de Crédito de fornecimento de bem móvel usado"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código da UF do emitente do Evento
```
```
P21 tpAutor E P17 N 1 - 1 1
```
```
Informar 5=Fisco.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco; 6=RFB; 8= Empresa
sucessora; 9=Outros Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 chaveNotaCredPres E P17 N 1 - 1 44 Chave da NFe com a indicação de crédito presumido apropriado
P24 tpNF E P17 N 1 - 1 1 Indicação de entrada ou saída da nota com o crédito presumido apropriado.
P25 gCredPres G P17 N 1 - N - Informações de crédito presumido apropriado na saída por item
P26 nitem A P25 N 1 - 1 1 - 3 Corresponde ao atributo “nItem” do elemento “det” do documento referenciado.
P27 pCredPres E P25 N 1 - 1 3v2- 4 Percentual do Crédito Presumido
P28 vCredPres E P25 N 1 - 1 13v2 Valor do Crédito Presumido
```
10.2.7.2. Leiaute Mensagem de Retorno

Retorno: Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral,

especificado no item 7.2.1.

10.2.7.3. Validação das Regras de Negócio **–** Específicas

Serão aplicadas as regras de validação gerais apresentadas no item Erro! Fonte de referência não encontrada. do MOC e as regras de negócio

específicas listadas a seguir.

```
# Regra de Validação Aplic. Msg Descrição Erro
Banco de Dados: NF-e
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

### P24- 10

```
Acessar BD e verificar se número do item do evento (tag:
gCredPres/nItem) informado existe na NFe referenciada
(tag: det/nItem)
```
```
Obrig. 1096 Rejeição: Número de item não existe na NFe
```
10.2.8. Evento: Solicitação de Apropriação de Crédito para bens e serviços que dependem de atividade do adquirente

Função: Evento a ser gerado pelo adquirente para apropriação de crédito de bens e serviços que dependam da sua atividade

Modelo: NF-e modelo 55

Autor do Evento: Destinatário da NFe (adquirente). A mensagem XML do evento será assinada com o certificado digital que tenha o CNPJ-Base (8

primeiras posições do CNPJ) ou CPF do Destinatário da NF-e.

Código do Tipo de Evento: 211150

10.2.8.1. Leiaute Mensagem de Entrada

Entrada: Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos

especificada na seção Erro! Fonte de referência não encontrada. do MOC.

Schema XML: envEventoNFe_v9.99.xsd

Schema XML - parte específica: e211150_v1.00.xsd

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 - Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 98 Descrição do evento: "Solicitação de Apropriação de Crédito para bens e serviços que dependem de atividade do adquirente"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código da UF do emitente do Evento
```
```
P21 tpAutor E P17 N 1 - 1 1
```
```
Informar 2=Empresa destinatária.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco; 6=RFB; 8= Empresa sucessora; 9=Outros
Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 gCredito G P17 N 1 - N 1 Informações de crédito
P24 nitem A P23 N 1 - 1 1 Corresponde ao atributo “nItem” do elemento “det” do documento referenciado.
P25 vCredIBS E P23 N 1 - 1 13v2 Valor da solicitação de crédito a ser apropriado de IBS
P26 vCredCBS E P23 N 1 - 1 13v2 Valor da solicitação de crédito a ser apropriado de CBS
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

10.2.8.2. Leiaute Mensagem de Retorno

Retorno: Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral,

especificado no item 7.2.1.

10.2.8.3. Validação das Regras de Negócio **–** Específicas

Serão aplicadas as regras de validação gerais apresentadas no item Erro! Fonte de referência não encontrada. do MOC e as regras de negócio

específicas listadas a seguir.

```
# Regra de Validação Aplic. Msg Descrição Erro
Banco de Dados: NF-e
```
### 2P21- 10

```
Se tpAutor=2-Empresa Destinatário:
```
- CNPJ/CPF do Autor diverge do CNPJ/CPF do Destinatário da
NF-e

```
Obrig. 575 Rejeição: Autor do evento diverge do destinatário da NF-e
```
### P24- 10

```
Acessar BD e verificar se número do item do evento (tag:
gCredito/nItem) informado existe na NFe referenciada (tag:
det/nItem)
```
```
Obrig. 1096 Rejeição: Número de item não existe na NFe
```
### P25- 10

```
Acessar BD e verificar se valor do crédito de IBS do evento (tag:
gCredito/vCredIBS) é maior que o valor do IBS do item informado
na NFe
```
```
Obrig. 1097
Rejeição: O valor do crédito de IBS do item não pode ser
maior que o valor do IBS do respectivo item na NFe.
```
### P26- 10

```
Acessar BD e verificar se valor do crédito de CBS do evento (tag:
gCredito/vCredCBS) é maior que o valor do IBS do item informado
na NFe
```
```
Obrig. 1098
Rejeição: O valor do crédito de CBS do item não pode ser
maior que o valor da CBS do respectivo item na NFe.
```
10.2.9. Manifestação sobre Pedido de Transferência de Crédito de IBS em Operações de Sucessão

Função: Evento a ser gerado pela sucessora em relação às notas fiscais de transferência de crédito de outra sucessora da mesma empresa sucedida

para informar aceite da transferência de crédito de IBS.

Autor: Empresa sucessora

Modelo: NF-e modelo 55

Código do Tipo de Evento: 212110


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

10.2.9.1. Leiaute Mensagem de Entrada

Entrada: Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos

especificada na seção Erro! Fonte de referência não encontrada. do MOC.


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

Schema XML: envEventoNFe_v9.99.xsd

Schema XML - parte específica: e212110.00.xsd

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 - 1 - 1 - Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
```
```
P19 descEvento E P17 C 1 - 1 85
Descrição do evento: "Manifestação sobre Pedido de Transferência de Crédito de IBS em Operações de
Sucessão"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código da UF do emitente do Evento
```
```
P21 tpAutor E P17 N 1 - 1 1
```
```
Informar 8=Empresa sucessora.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco; 6=RFB; 8= Empresa sucessora;
9=Outros Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
```
```
P23 indAceitacao A P17 N 1 - 1 1
Indicador de aceitação do valor de transferência para a empresa que emitiu a nota referenciada. Valores: 0 = não
aceite; 1 = aceite.
```
10.2.9.2. Leiaute Mensagem de Retorno

Retorno: Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral,

especificado no item 7.2.1.

10.2.10. Manifestação sobre Pedido de Transferência de Crédito CBS em Operações de Sucessão

Função: Evento a ser gerado pela sucessora em relação às notas fiscais de transferência de crédito de outra sucessora da mesma empresa sucedida

para informar aceite da transferência de crédito de CBS.

Autor: Empresa sucessora

Modelo: NF-e modelo 55

Código do Tipo de Evento: 212120

10.2.10.1. Leiaute Mensagem de Entrada

Entrada: Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos

especificada na seção Erro! Fonte de referência não encontrada. do MOC.


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

Schema XML: envEventoNFe_v9.99.xsd

Schema XML - parte específica: e 212120 _v1.00.xsd

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 - Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
```
```
P19 descEvento E P17 C 1 - 1 82
Descrição do evento: "Manifestação sobre Pedido de Transferência de Crédito CBS em
Operações de Sucessão"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código da UF do emitente do Evento
```
```
P21 tpAutor E P17 N 1 - 1 1
```
```
Informar 8=Empresa sucessora.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco; 6=RFB; 8=
Empresa sucessora; 9=Outros Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
```
```
P23 indAceitacao A P17 N 1 - 1 1
Indicador de aceitação do valor de transferência para a empresa que emitiu a nota referenciada.
Valores: 0 = não aceite; 1 = aceite.
```
10.2.10.2. Leiaute Mensagem de Retorno

Retorno: Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral,

especificado no item 7.2.1.

10.2.11. Evento: Manifestação do Fisco sobre Pedido de Transferência de Crédito de IBS em Operações de Sucessão

Função: Evento a ser gerado pelo fisco em relação às notas fiscais de transferência de crédito para informar aceite ou não aceite da transferência de

crédito de IBS.

Autor: Fisco

Modelo: NF-e modelo 55

Código do Tipo de Evento: 412120

10.2.11.1. Leiaute Mensagem de Entrada

Entrada: Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos

especificada na seção Erro! Fonte de referência não encontrada. do MOC.


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

Schema XML: envEventoNFe_v9.99.xsd

Schema XML - parte específica: e 412120 _v1.00.xsd

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 - Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
```
```
P19 descEvento E P17 C 1 - 1 94
Descrição do evento: "Manifestação do Fisco sobre Pedido de Transferência de Crédito de IBS em
Operações de Sucessão"
P20 cOrgaoAutor E P17 N 1 - 1 2 Código da UF do emitente do Evento
```
```
P21 tpAutor E P17 N 1 - 1 1
```
```
Informar 5=Fisco.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco; 6=RFB; 8= Empresa
sucessora; 9=Outros Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
```
```
P23 indDeferimento A P17 N 1 - 1 1
Indicador de aceitação do valor de transferência para a empresa que emitiu a nota referenciada. Valores: 0 =
não aceite; 1 = aceite.
P24 cMotivo E P17 N 1 - 1 1 1 – Falta de manifestação de todas as sucessoras. 2 – Outros
P25 xMotivo E P17 C 1 - 1 500
```
10.2.11.2. Leiaute Mensagem de Retorno

Retorno: Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral,

especificado no item 7.2.1.

10.2.12. Manifestação do Fisco sobre Pedido de Transferência de Crédito de CBS em Operações de Sucessão

Função: Evento a ser gerado pelo fisco em relação às notas fiscais de transferência de crédito para informar aceite ou não aceite da transferência de

crédito de CBS.

Autor: Fisco

Modelo: NF-e modelo 55

Código do Tipo de Evento: 412130

10.2.12.1. Leiaute Mensagem de Entrada


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

Entrada: Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos

especificada na seção Erro! Fonte de referência não encontrada. do MOC.

Schema XML: envEventoNFe_v9.99.xsd

Schema XML - parte específica: e 412130 _v1.00.xsd

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 1 - 1 - Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 94
P20 cOrgaoAutor E P17 N 1 - 1 2
Descrição do evento: "Manifestação do Fisco sobre Pedido de Transferência de Crédito de CBS em
Operações de Sucessão"
```
```
P21 tpAutor E P17 N 1 - 1 1
```
```
Informar 5=Fisco.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco; 6=RFB; 8= Empresa
sucessora; 9=Outros Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
```
```
P23 indDeferimento A P17 N 1 - 1 1
Indicador de aceitação do valor de transferência para a empresa que emitiu a nota referenciada.
Valores: 0 = não aceite; 1 = aceite.
P24 cMotivo E P17 N 1 - 1 1 1 – falta de manifestação de todas as sucessoras. 2 – Outros
P25 xMotivo E P17 C 1 - 1 500
```
10.2.12.2. Leiaute Mensagem de Retorno

Retorno: Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral,

especificado no item 7.2.1.

10.2.13. Evento: Evento de cancelamento de evento do emitente

Função: Permitir que o emitente da NFe efetue o cancelamento de evento de NFe.

Modelo: NF-e modelo 55

Autor do Evento: Emitente da NFe (emitente)

Tipo de Evento (Código - Descrição):

112111 - Cancelamento do evento 112110


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

10.2.13.1. Leiaute Mensagem de Entrada

Entrada: Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos

especificada na seção Erro! Fonte de referência não encontrada. do MOC.

Schema XML: envEventoNFe_v9.99.xsd

Schema XML - parte específica: e112XXX_v1.00.xsd

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06 - 1 - 1 - Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 29 Descrição do evento.
P20 cOrgaoAutor E P17 N 1 - 1 2 Código do Órgão Autor do Evento. Informar o Código da UF para este Evento.
P21 tpAutor E P17 N 1 - 1 1 Informar 1=Empresa emitente.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco;
6=RFB; 9=Outros Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 nProtEvento E P17 N 1 - 1 1 Informar o número do Protocolo de Autorização do Evento da
NF-e a que se refere este cancelamento.
```
10.2.13.2. Leiaute Mensagem de Retorno

Retorno: Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral,

especificado no item 7.2.1.

10.2.13.3. Validação das Regras de Negócio **–** Específicas

Serão aplicadas as regras de validação gerais apresentadas no item Erro! Fonte de referência não encontrada. do MOC e as regras de negócio

específicas listadas a seguir.

```
# Regra de Validação Aplic. Msg Descrição Erro
*** Banco de Dados: Evento 2
```
```
4P15- 10
Acesso BD de Eventos (Chave: Chave de Acesso,
tpEvento=XXXXXX, nSeqEvento): - Evento inexistente
Obrig. 459 Rejeição: Cancelamento de Evento inexistente
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

```
4P23- 10 Número do Protocolo não encontrado Obrig. 460 Rejeição: Protocolo do Evento difere do cadastrado
```
10.2.14. Evento: Evento de cancelamento de evento do destinatário

Função: Permitir que o destinatário da NFe efetue o cancelamento de evento de NFe.

Modelo: NF-e modelo 55

Autor do Evento: Destinatário da NF-e. A mensagem XML do evento será assinada com o certificado digital que tenha o CNPJ-Base (8 primeiras

posições do CNPJ) ou CPF do Destinatário da NF-e.

Tipo de Evento (Código - Descrição):

- 211111 - Cancelamento do evento 211110
- 211121 - Cancelamento do evento 211120
- 211131 - Cancelamento do evento 211130
- 211141 - Cancelamento do evento 211140
- 211151 - Cancelamento do evento 211150

10.2.14.1. Leiaute Mensagem de Entrada

Entrada: Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos

especificada na seção Erro! Fonte de referência não encontrada. do MOC.

Schema XML: envEventoNFe_v9.99.xsd

Schema XML - parte específica: e211XXX_v1.00.xsd

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06
1 - 1 - Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 29 Descrição do evento.
P20 cOrgaoAutor E P17 N 1 - 1 2 Código do Órgão Autor do Evento. Informar o Código da UF para este Evento.
P21 tpAutor E P17 N 1 - 1 1 Informar 2=Empresa destinatária.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco;
6=RFB; 8= Empresa sucessora; 9=Outros Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 nProtEvento E P17 N 1 - 1 1 Informar o número do Protocolo de Autorização do Evento da
NF-e a que se refere este cancelamento.
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

10.2.14.2. Leiaute Mensagem de Retorno

Retorno: Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral,

especificado no item 7.2.1.

10.2.14.3. Validação das Regras de Negócio **–** Específicas

Serão aplicadas as regras de validação gerais apresentadas no item Erro! Fonte de referência não encontrada. do MOC e as regras de negócio

específicas listadas a seguir.

```
# Regra de Validação Aplic. Msg Descrição Erro
```
## Banco de Dados: NF-e

### 2P21- 10

```
Se tpAutor=2-Empresa Destinatário:
```
- CNPJ/CPF do Autor diverge do
CNPJ/CPF do Destinatário da NF-e

```
Obrig. 575 Rejeição: Autor do evento diverge do destinatário da NF-e
```
```
Banco de Dados: Evento 2
```
### 4P15- 10

```
Acesso BD de Eventos (Chave: Chave de
Acesso, tpEvento=XXXXXX, nSeqEvento): -
Evento inexistente
```
```
Obrig. 459 Rejeição: Cancelamento de Evento inexistente
```
```
4P23- 10 Número do Protocolo não encontrado Obrig. 460 Rejeição: Protocolo do Evento difere do cadastrado
```
10.2.15. Evento: Evento de cancelamento de evento da sucessora

Função: Permitir que a sucessora efetue o cancelamento de evento de NFe.

Modelo: NF-e modelo 55

Autor do Evento: Destinatário da NF-e. A mensagem XML do evento será assinada com o certificado digital que tenha o CNPJ-Base (8 primeiras

posições do CNPJ) ou CPF do Destinatário da NF-e.

Tipo de Evento (Código - Descrição):


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

- 212111 - Cancelamento do evento 212110
- 212121 - Cancelamento do evento 212120

10.2.15.1. Leiaute Mensagem de Entrada

Entrada: Estrutura XML da parte específica do evento, a ser inserida na tag detEvento (P17) da Parte Geral do Web Service de Registro de Eventos

especificada na seção Erro! Fonte de referência não encontrada. do MOC.

Schema XML: envEventoNFe_v9.99.xsd

Schema XML - parte específica: e212XXX_v1.00.xsd

```
# Campo Ele Pai Tipo Ocor. Tam. Descrição/Observação
P17 detEvento G P06
1 - 1 - Detalhes do Evento
P18 versao A P17 N 1 - 1 2v2 Versão do leiaute do evento (P16)
P19 descEvento E P17 C 1 - 1 29 Descrição do evento.
P20 cOrgaoAutor E P17 N 1 - 1 2 Código do Órgão Autor do Evento. Informar o Código da UF para este Evento.
P21 tpAutor E P17 N 1 - 1 1 Informar 8=Empresa sucessora.
Valores: 1=Empresa Emitente, 2=Empresa destinatária; 3=Empresa; 5=Fisco;
6=RFB; 8= Empresa sucessora; 9=Outros Órgãos.
P22 verAplic E P17 N 1 - 1 1 - 20 Versão do aplicativo do autor do evento.
P23 nProtEvento E P17 N 1 - 1 1 Informar o número do Protocolo de Autorização do Evento da
NF-e a que se refere este cancelamento.
```
10.2.15.2. Leiaute Mensagem de Retorno

Retorno: Estrutura XML com a mensagem do resultado da transmissão, conforme retorno do Web Service de Registro de Eventos – Parte Geral,

especificado no item 7.2.1.

10.2.15.3. Validação das Regras de Negócio **–** Específicas

Serão aplicadas as regras de validação gerais apresentadas no item Erro! Fonte de referência não encontrada. do MOC e as regras de negócio

específicas listadas a seguir.

```
# Regra de Validação Aplic. Msg Descrição Erro
```

## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

```
Banco de Dados: Evento 2
```
```
4P15- 10
```
```
Acesso BD de Eventos (Chave: Chave de Acesso,
tpEvento=XXXXXX, nSeqEvento): - Evento inexistente
Obrig. 459 Rejeição: Cancelamento de Evento inexistente
```
```
4P23- 10 Número do Protocolo não encontrado Obrig. 460 Rejeição: Protocolo do Evento difere do cadastrado
```
11. DANFE

Alterações no DANFE para exibir informações relativas aos novos tributos estão em estudo, e serão

publicadas em uma nova versão desta Nota Técnica.


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

ANEXO I

NCM DO IMPOSTO SELETIVO

Link para a tabela:

https://docs.google.com/spreadsheets/d/1TnXQPmAgAyvOgSIznmw1oxdVmvUIqkYTMH0Xxbepby8/edit?usp=sharing

ANEXO II

CÓDIGO DE CLASSIFICAÇÃO TRIBUTÁRIA DO IMPOSTO SELETIVO

Tabela a ser publicada.

ANEXO III

CÓDIGO DE CLASSIFICAÇÃO TRIBUTÁRIA DO IBS E DA CBS

Tabela a ser publicada.


## Reforma Tributária - Emenda Constitucional 132/2023

NT 2024.002 Versão 1. 10

ANEXO IV

CÓDIGO DE CLASSIFICAÇÃO DO CRÉDITO PRESUMIDO

Tabela a ser publicada.
