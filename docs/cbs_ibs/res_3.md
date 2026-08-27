# Projeto Reforma Tributária do

# Consumo – Adequações NF-e / NFC-e

## Nota Técnica 2024.0 02 – Versão 1.

## 01 de agosto de 2024


## Reforma Tributária do Consumo

Sumário

   - NT-RT 2024.00 2 Versão 1.
- Introdução Contents
- 1 Tipos Básicos da Tributação
- 2 Código Situação Tributária e Código de Classificação da Tributação
- 3 Finalidades débito e crédito da NF-e
- 4 Alterações no arquivo XML da NF-e
      - 4.1 Esquema gráfico do leiaute do IBS da UF, contemplando IBS, CBS e Imposto Seletivo.
      - 4.2 Inserção dos novos campos de informação no leiaute da NF-e e NFC-e
      - 4.4 Disposição dos campos do Imposto Seletivo
      - 4.5 Imposto de Bens e Serviços - IBS das UFs e Municípios
      - 4.6 Disposição dos campos do IBS da UF
      - 4.7 Disposição dos campos do IBS dos Municípios
      - 4.9 Disposição dos campos da CBS
      - 4.10 Disposição dos campos do IBS Monofásico
      - 4.11 Disposição dos campos totalizadores
- 5 Leiaute da NF-e (Modelo 55 e 65)
      - 5.1 Grupo UB. Informações dos tributos IBS / CBS e Imposto Seletivo
      - 5.2 Grupo W. Total da NF-e
- 6 Regras de Validação
      - 6.1 Grupo UB. Informações dos tributos IBS / CBS e Imposto Seletivo
- 7 Eventos
         - Franca de Manaus - ZFM” 7.1 Evento: “Decurso de Prazo de Internalização na Área de Livre Comércio - ALC ou Zona
         - Comércio - ALC ou Zona Franca de Manaus - ZFM” 7.2 Evento: “Cancelamento do Evento Decurso de Prazo de Internalização na Área de Livre
      - 7.3 Evento: “Solicitação de Apropriação de Crédito Presumido”
      - 7.4 Evento: “Cancelamento do Evento Solicitação de Apropriação de Crédito Presumido”
      - 7.5 Evento: “Destinação de Item para Consumo Pessoal”
      - 7.6 Evento: “Cancelamento do Evento Destinação de Item para Consumo Pessoal”
      - 7.7 Evento: “Imobilização de Item”
      - 7.8 Evento: “Cancelamento do Evento Imobilização de Item”
      - 7.9 Evento: “Solicitação de Apropriação de Crédito de Combustível”..................................
         - 7.10 Evento: “Cancelamento do Evento Solicitação de Apropriação de Crédito de Combustível”
- 8 DANFE
- ANEXO I - NCM DO IMPOSTO SELETIVO
- ANEXO II - CÓDIGO DE CLASSIFICAÇÃO TRIBUTÁRIA DO IMPOSTO SELETIVO
- ANEXO III - CÓDIGO DE CLASSIFICAÇÃO TRIBUTÁRIA DO IBS E DA CBS


```
Reforma Tributária do Consumo
```
### NT-RT 2024.00 2 Versão 1.

Controle de Versões

```
Versão Publicação Descrição
```
```
1.00 08 /202 4 NT que inseri campos e regras para informação do Imposto de Bens e Serviços - IBS,
Contribuição de Bens e Serviços - CBS e Imposto Seletivo - IS.
```
Histórico de Alterações / Cronograma

```
Versão Histórico de atualizações Implantação
Teste
```
```
Implantação
Produção
```
```
1.00 Versão inicial com a previsIBS e CBS na NF-e e NFC-e.ão dos campos do Imposto Seletivo, 01/09/2025 31/ 10 //20 25
```

```
Reforma Tributária do Consumo
NT-RT 2024.00 2 Versão 1.
```
Introdução

O Projeto de Lei Complementar Federal PLP 68 , aprovado na Câmara dos Deputados e
encaminhado para aprovação junto ao Senado Federal, definiu na Seção VIII – Disposições
Transitórias, Art. 61, a obrigatoriedade para Estados, o Distrito Federal e os Municípios adaptarem
os sistemas autorizadores de Documentos Fiscais Eletrônicos (DFe) vigentes para utilização de
leiaute padronizado, que permita aos contribuintes informarem os dados relativos ao Imposto sobre
Bens e Serviços (IBS), Contribuição sobre Bens e Serviços (CBS) e Imposto Seletivo (IS).

Como as infraestruturas autorizadoras de Documentos Fiscais Eletrônicos (DFe) das Unidades
Federadas e Municípios, além das aplicações e sistemas dos contribuintes, necessitam de, no
mínimo, 1 ano para o desenvolvimento das alterações necessárias, estamos divulgando esta Nota
Técnica (NT) para implantação, em produção, a partir do dia 31/ 10 /2025, de forma a entrar em
efetiva operacionalização a partir 01/01/2026.

Como as discussões envolvendo a implantação da Reforma Tributária ainda estão em curso,
esclarecemos que esta NT poderá ser ajustada ao longo do seu processo de execução, da mesma
forma como ocorre com as demais NT já implementadas.

Desta forma, esta NT, discutida conjuntamente com a RFB e entidades representantes dos
municípios, modifica o leiaute da NF-e e NFC-e, inserindo os grupos e campos opcionais
relacionados a tributação do Imposto sobre Bens e Serviços (IBS), Contribuição sobre Bens e
Serviços (CBS) e Imposto Seletivo (IS), em atendimento as alterações previstas na Emenda
Constitucional 132 de 20 de dezembro de 2023 para implementação da Reforma Tributária.

1 Tipos Básicos da Tributação

Em busca de uma padronização entre os diversos documentos fiscais eletrônicos existentes, esta
NT introduz o arquivo “DFeTiposBasicos_v1.00.xsd” ao conjunto dos arquivos que compõem do
schema de todos os Documentos Fiscais Eletrônicos - DF-e, entre eles a NF-e e NFC-e.

Este arquivo define de forma estruturada a previsão de campos a serem informados para o registro
das informações referentes a tributação do IBS e da CBS em um tipo complexo referenciado no
leiaute padrão da NF-e e NFC-e conforme estrutura demonstrada no item 5.1, e também será
utilizado nos demais documentos fiscais eletrônicos.

2 Código Situação Tributária e Código de Classificação da

Tributação

O grupo de informações do IBS, CBS e IS associado aos itens do documento fiscal devem ser
classificados de acordo com Código de Situação Tributária (CST) e Código de Classificação
Tributária (cClassTrib) do IBS, CBS e IS.

As duas tabelas estarão publicadas nos portais nacionais dos documentos fiscais eletrônicos e as
ocorrências previstas para preenchimento estão relacionadas às previsões legais do texto da
Reforma Tributária.


```
Reforma Tributária do Consumo
NT-RT 2024.00 2 Versão 1.
Associados a estas tabelas, serão publicados indicadores de obrigatoriedade de preenchimento dos
grupos de informações do IBS, da CBS e do IS nos itens da DF-e. Essas configurações vinculam
de forma dinâmica os CST e cClassTrib com as Regras de Validação descritas logo abaixo.
```
3 Finalidades débito e crédito da NF-e

```
Notas de Débito e Crédito são nomes de instrumentos utilizados mundialmente para documentar
situações contábeis onde é necessário corrigir informações comerciais que foram registradas em
um documento, que no Brasil é a Nota Fiscal.
```
```
Esta Nota Técnica cria na NF-e modelo 55 as finalidades de emissão correspondentes. O sentido
das palavras “débito” e “crédito” sempre se referem ao ponto de vista do emissor:
```
- Uma nota de débito documenta uma situação na qual o emitente registra um aumento no imposto
    devido (consequentemente, uma redução no imposto devido pelo adquirente, que é o
    destinatário);
- Uma nota de crédito documenta uma situação na qual o emitente registra uma redução no
    imposto devido (consequentemente, um aumento no imposto devido pelo adquirente, que é o
    destinatário);

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
eletrônicos, em cumprimento ao que preconiza o PLP 68. A menos que ocorra alteração na
regulamentação do ICMS e do IPI, notas de crédito e notas de débito não poderão ser utilizadas
para ajustes relativos a estes tributos.
```

```
Reforma Tributária do Consumo
NT-RT 2024.00 2 Versão 1.
```
4 Alterações no arquivo XML da NF-e

4.1 Esquema gráfico do leiaute do IBS da UF, contemplando IBS, CBS e

Imposto Seletivo.


```
Reforma Tributária do Consumo
NT-RT 2024.00 2 Versão 1.
```
4.2 Inserção dos novos campos de informação no leiaute da NF-e e

NFC-e


```
Reforma Tributária do Consumo
NT-RT 2024.00 2 Versão 1.
```
4.4 Disposição dos campos do Imposto Seletivo


```
Reforma Tributária do Consumo
NT-RT 2024.00 2 Versão 1.
```
4.5 Imposto de Bens e Serviços - IBS das UFs e Municípios

4.6 Disposição dos campos do IBS da UF


```
Reforma Tributária do Consumo
NT-RT 2024.00 2 Versão 1.
```
4.7 Disposição dos campos do IBS dos Municípios


```
Reforma Tributária do Consumo
NT-RT 2024.00 2 Versão 1.
```
4.9 Disposição dos campos da CBS


```
Reforma Tributária do Consumo
NT-RT 2024.00 2 Versão 1.
```
4.10 Disposição dos campos do IBS Monofásico


```
Reforma Tributária do Consumo
NT-RT 2024.00 2 Versão 1.
```
4.11 Disposição dos campos totalizadores


```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.
```
5 Leiaute da NF-e (Modelo 55 e 65)

Grupo B. Identificação da Nota Fiscal eletrônica
**# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
29 B25 finNFe Finalidade de emissão da NF-e E B01 N 1 - 1 1 1=NF-e normal;
2=NF-e complementar;
3=NF-e de ajuste;
4=Devolução de mercadoria.
5=Nota de crédito
6=Nota de débito**

Grupo N01. ICMS Normal e ST
**# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
164 N01 ICMS Informações do ICMS da Operação própria e ST CG M01 0 - 1 Informar apenas um dos grupos de tributação do ICMS
(ICMS00, ICMS10, ...) (v2.0)**

5.1 Grupo UB. Informações dos tributos IBS / CBS e Imposto Seletivo

```
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
```
```
324.01 UB01 IBSCBSSel Informações do Imposto de Bens e Serviços -^ IBS,
Contribuição de Bens e Serviços - CBS e Imposto Seletivo
```
```
H01^^0 -^1^
```
```
324.02 UB02 seletivo Informações do Imposto Seletivo G UB04 0 - 1
```
```
324.03 UB03 CST Código de Situação Tributária do Imposto Seletivo E UB02 N 1 - 1
```
```
324.04 UB04 cClassTrib Código de Classificação Tributária do Imposto Seletivo E UB05 C 1 - 1
```
```
324.05 UB05 gImpSel Grupo de Informações do Imposto Seletivo G UB02 0 - 1
```
```
324.06 UB06 vBCImpSel Valor da Base de Cálculo do Imposto Seletivo E UB05 N 1 - 1 13v
```

```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.
```
```
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
```
**324.07 UB07 pImpSel Alíquota do Imposto Seletivo E UB05 N 1 - 1 3v2- 4**

**324.08 UB08 pImpSelEspec Alíquota específica por unidade de medida apropriada E UB05 N 0 - 1 3v2- 4**

**324.09 UB09 uTrib Unidade de Medida Tributável E UB05 C 0 - 1 1 - 6**

**324.10 UB10 qTrib Quantidade Tributável E UB05 N 0 - 1 11v0- 4**

**324.11 UB11 vImpSel Valor do Imposto Seletivo E UB05 N 1 - 1 13v**

**324.12 UB12 CST Código de Situação Tributária do IBS e CBS E UB01 N 0 - 1**

**324.13 UB13 cClassTrib Código de Classificação Tributária do IBS e CBS E UB01 C 0 - 1**

**324.14 UB14 gIBSCBS Grupo de Informações do IBS, CBS e Imposto Seletivo CG UB01 1 - 1**

**324.15 UB15 vBC Base de cálculo do IBS e CBS E UB04 N 1 - 1 3v2- 4**

**324.16 UB16 gIBSUF Grupo de Informações do IBS para a UF G UB14 1 - 1**

**324.17 UB17 pIBSUF Alíquota do IBS de competência das UF E UB15 N 1 - 1 3v2- 4 Alíquota vigente do IBS da UF**

**324.18 UB18 - x- Sequencia XML G UB15 0 - 1**

**324.19 UB19 vTribOP Valor bruto do tributo na operação E UB17 N 1 - 1 13v**

```
Valor do tributo considerando BC x Alq do IBS, sem considerar
qualquer desoneração.
```
**324.20 UB20 gCredPres Grupo de Informações do Crédito Presumido G UB17 0 - 1** (^)
**Grupo de Informações do Crédito Presumido, quando
aproveitado pelo emitente do documento. Exemplos: 1 -
Aquisição de PR não contribuinte. 2 - Tomador de serviço de
transporte de TAC PF não contrib. 3 - Aquisição de pessoa
física com destino a reciclagem. 4 - Aquisição de bens móveis
de PF não contrib. para revenda (veículos / brecho). 5 -
Regime opcional para cooperativa.
324.21 UB21 pCredPres Percentual do Crédito Presumido E UB19 N 1 - 1 3v2- 4
324.22 UB22 vCredPres Valor do Crédito Presumido E UB19 N 1 - 1 13v
324.23 UB23 gDif Grupo de Informações do Diferimento G UB17 0 - 1
324.24 UB24 pDif Percentual do diferimento E UB22 N 1 - 1 3v2- 4
324.25 UB25 vDif Valor do Diferimento E UB22 N 1 - 1 13v**


```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.
```
```
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
```
**324.26 UB26 gDevTrib Grupo de Informações da devolução de tributos G UB17 0 - 1**

**324.27 UB27 vDevTrib Valor do tributo devolvido E UB25 N 1 - 1 13v**

```
Valor do tributo devolvido. No fornecimento de energia
elétrica, água, esgoto e gás natural e em outras hipóteses
definidas no regulamento
```
**324.28 UB28 gRed Grupo de informações da redução da alíquota G UB15 0 - 1**

**324.29 UB29 pRedAliq Percentual da redução de alíquota E UB27 N 1 - 1 3v2- 4**

**324.30 UB30 pAliqEfet**

```
Aliquota Efetiva do IBS de competência das UF que será
aplicada a Base de Cálculo E UB27 N 1 - 1 3v2- 4 Alíquota efetiva, após aplicação da redução de alíquota.
```
**324.31 UB31 gDeson Grupo de informações da Desoneração G UB15 0 - 1** (^)
**Grupo de informações da Desoneração. Exemplo 1: Art. 442,
§4. Operações com ZFM e ALC. Exemplo 2: Operações com
suspensão do tributo.
324.32 UB32 CST Código de Situação Tributária do IBS e CBS E UB30 N 0 - 1 Informado como se a operação fosse tributada integralmente
324.33 UB33 cClassTrib Código de Classificação Tributária do IBS e CBS E UB30 C 0 - 1 Informado como se a operação fosse tributada integralmente
324.34 UB34 vBC Valor da BC E UB30 N 0 - 1 13v2 Info: Avaliando retirar e deixar a BC do grupo anterior
324.35 UB35 pAliq Valor da alíquota E UB30 N 1 - 1 3v2- 4 Informado como se a operação fosse tributada integralmente
324.36 UB36 vDeson Valor desonerado E UB30 N 1 - 1 13v
324.37 UB37 vIBSUF Valor do IBS de competência da UF E UB15 N 1 - 1 13v
324.38 UB38 gIBSMun Grupo de Informações do IBS para o município G UB14 1 - 1
324.39 UB39 pIBSMun Alíquota do IBS de competência dos Municípios E UB37 N 1 - 1 3v2- 4
324.40 UB40 - x- Sequencia XML G UB37 0 - 1
324.41 UB41 vTribOP Valor bruto do tributo na operação E UB39 N 1 - 1 13v
Valor do tributo considerando BC x Alq do IBS, sem considerar
qualquer desoneração.**


```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.
```
```
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
```
**324.42 UB42 gCredPres Grupo de Informações do Crédito Presumido G UB39 0 - 1** (^)
**Grupo de Informações do Crédito Presumido, quando
aproveitado pelo emitente do documento. Exemplos: 1 -
Aquisição de PR não contribuinte. 2 - Tomador de serviço de
transporte de TAC PF não contrib. 3 - Aquisição de pessoa
física com destino a reciclagem. 4 - Aquisição de bens móveis
de PF não contrib. para revenda (veículos / brecho). 5 -
Regime opcional para cooperativa.
324.43 UB43 pCredPres Percentual do Crédito Presumido E UB41 N 1 - 1 3v2- 4
324.44 UB44 vCredPres Valor do Crédito Presumido E UB41 N 1 - 1 13v
324.45 UB45 gDif Grupo de Informações do Diferimento G UB39 0 - 1
324.46 UB46 pDif Percentual do diferimento E UB43 N 1 - 1 3v2- 4
324.47 UB47 vDif Valor do Diferimento E UB43 N 1 - 1 13v
324.48 UB48 gDevTrib Grupo de Informações da devolução de tributos G UB39 0 - 1
324.49 UB49 vDevTrib Valor do tributo devolvido E UB46 N 1 - 1 13v
Valor do tributo devolvido. No fornecimento de energia
elétrica, água, esgoto e gás natural e em outras hipóteses
definidas no regulamento
324.50 UB50 gRed Grupo de informações da redução da alíquota G UB37 0 - 1
324.51 UB51 pRedAliq Percentual da redução de alíquota E UB48 N 1 - 1 3v2- 4
324.52 UB52 pAliqEfet
Aliquota Efetiva do IBS de competência dos Municípios
que será aplicada a Base de Cálculo E UB48 N 1 - 1 3v2- 4 Alíquota efetiva, após aplicação da redução de alíquota.
324.53 UB53 gDeson Grupo de informações da Desoneração G UB37 0 - 1
Grupo de informações da Desoneração. Exemplo 1: Art. 442,
§4. Operações com ZFM e ALC. Exemplo 2: Operações com
suspensão do tributo.
324.54 UB54 CST Código de Situação Tributária do IBS e CBS E UB51 N 0 - 1 Informado como se a operação fosse tributada integralmente
324.55 UB55 cClassTrib Código de Classificação Tributária do IBS e CBS E UB51 C 0 - 1 Informado como se a operação fosse tributada integralmente
324.56 UB56 vBC Valor da BC E UB51 N 1 - 1 13v2 Info: Avaliando retirar e deixar a BC do grupo anterior
324.57 UB57 pAliq Valor da alíquota E UB51 N 1 - 1 3v2- 4 Informado como se a operação fosse tributada integralmente
324.58 UB58 vDeson Valor desonerado E UB51 N 1 - 1 13v**


```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.
```
```
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
```
**324.59 UB59 vIBSMun Valor do IBS de competência dos Municípios E UB37 N 1 - 1 13v**

**324.60 UB60 gCBS Grupo de Informações da CBS G UB14 1 - 1**

**324.61 UB61 pCBS Alíquota da CBS E UB58 N 1 - 1 3v2- 4**

**324.62 UB62 - x- Sequencia XML G UB58 0 - 1**

**324.63 UB63 vTribOp Valor bruto do tributo na operação E UB60 N 1 - 1**

(^) **Valor do tributo considerando BC x Alq da CBS, sem
considerar qualquer desoneração.
324.64 UB64 gCredPres Grupo de Informações do Crédito Presumido G UB60 0 - 1** (^)
**Grupo de Informações do Crédito Presumido, quando
aproveitado pelo emitente do documento. Exemplos: 1 -
Aquisição de PR não contribuinte. 2 - Tomador de serviço de
transporte de TAC PF não contrib. 3 - Aquisição de pessoa
física com destino a reciclagem. 4 - Aquisição de bens móveis
de PF não contrib. para revenda (veículos / brecho). 5 -
Regime opcional para cooperativa.
324.65 UB65 pCredPres Percentual do Crédito Presumido E UB62 N 1 - 1 3v2- 4
324.66 UB66 vCredPres Valor do Crédito Presumido E UB62 N 1 - 1 13v
324.67 UB67 gDif Grupo de Informações do Diferimento G UB60 0 - 1
324.68 UB68 pDif Percentual do diferimento E UB65 N 1 - 1 3v2- 4
324.69 UB69 vDif Valor do Diferimento E UB65 N 1 - 1 13v
324.70 UB70 gDevTrib Grupo de Informações da devolução de tributos G UB60 0 - 1
324.71 UB71 vDevTrib Valor do tributo devolvido E UB68 N 1 - 1 13v
Valor do tributo devolvido. No fornecimento de energia
elétrica, água, esgoto e gás natural e em outras hipóteses
definidas no regulamento
324.72 UB72 gRed Grupo de informações da redução da alíquota G UB58 0 - 1
324.73 UB73 pRedAliq Percentual da redução de alíquota E UB70 N 1 - 1 3v2- 4
324.74 UB74 pAliqEfet
Alíquota Efetiva da CBS que será aplicada a Base de
Cálculo E UB70 N 1 - 1 3v2- 4 Alíquota efetiva, após aplicação da redução de alíquota.
324.75 UB75 gDeson Grupo de informações da Desoneração G UB
0 - 1**
(^) **Grupo de informações da Desoneração. Exemplo 1: Art. 442,
§4. Operações com ZFM e ALC. Exemplo 2: Operações com
suspensão do tributo.**


```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.
```
```
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
```
**324.76 UB76 CST Código de Situação Tributária do IBS e CBS E UB73 N 0 - 1 Informado como se a operação fosse tributada integralmente**

**324.77 UB77 cClassTrib Código de Classificação Tributária do IBS e CBS E UB73 C 0 - 1 Informado como se a operação fosse tributada integralmente**

**324.78 UB78 vBC Valor da BC E UB73 N 1 - 1 13v2 Info: Avaliando retirar e deixar a BC do grupo anterior**

**324.79 UB79 pAliq Valor da alíquota E UB73 N 1 - 1 3v2- 4 Informado como se a operação fosse tributada integralmente**

**324.80 UB80 vDeson Valor desonerado E UB73 N 1 - 1 13v**

**324.81 UB81 vCBS Valor da CBS E UB58 N 1 - 1 13v**

**324.82 UB82 gIBSCBSMono**

```
Grupo de Informações do IBS e CBS em operações com
imposto monofásico CG UB
1 - 1
Monofasia dos Combustíveis
```
**324.83 UB83 qBCMono Quantidade tributada na monofasia E UB80 N 0 - 1 11v0- 4**

```
Informar a BC quantidade conforme unidade de medida
estabelecida na legislação para o produto.
```
**324.84 UB84 adRemIBS Alíquota ad rem do IBS E UB80 N 1 - 1 3v2- 4**

**324.85 UB85 adRemCBS Alíqutoa ad rem da CBS E UB80 N 1 - 1 3v2- 4**

**324.86 UB86 vIBSMono Valor do IBS monofásico E UB80 N 1 - 1 13v**

```
O valor do imposto é obtido pela multiplicação da alíquota ad
rem pela quantidade do produto conforme unidade de
medida estabelecida na legislação.
```
**324.87 UB87 vCBSMono Valor da CBS monofásica E UB80 N 1 - 1 13v**

```
O valor do imposto é obtido pela multiplicação da alíquota ad
rem pela quantidade do produto conforme unidade de
medida estabelecida na legislação.
```
**324.88 UB88 - x- Sequencia XML G UB80 0 - 1** (^)
**Uso em operações com combustíveis derivados de petróleo
(Gasolina A) [ou *Óleo Diesel A*] para retenção do imposto
sobre o biocombustível a ser misturado. Art 173 PLP 68/24.
324.89 UB89 qBCMonoReten Quantidade tributada sujeita à retenção na monofasia E UB84 N 0 - 1 11v0- 4
Informar a BC do ICMS sujeita a retenção em quantidade
conforme unidade de medida estabelecida na legislação para
o produto.
324.90 UB90 adRemIBSREten Alíquota ad rem do imposto sujeito a retenção E UB84 N 1 - 1 3v2-**^4^
**324.91 UB91 vIBSMonoReten Valor do IBS monofásico sujeito a retenção E UB84 N 1 - 1 13v2 Valor do IBS com retenção, a ser somado ao valor de IBS a ser recolhido.**


```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.
```
```
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
```
**324.92 UB92 - x- Sequencia XML G UB80 0 - 1**

```
Uso em operações com crédito presumido. Exemplo: compra
de combustível por transportador de passageiros, venda para
órgão público, venda equiparada a exportação, entre outros.
```
**324.93 UB93 pCredPresIBS Percentual de crédito presumido do IBS monofásico. E UB88 N 1 - 1 3v2- 4**

**324.94 UB94 vCRedPresIBS Valor do crédito presumido do IBS monofásico. E UB88 N 1 - 1 13v**^

**324.95 UB95 pCredPresCBS Percentual de crédito presumido da CBS monofásica. E UB88 N 1 - 1 3v2- 4**

**324.96 UB96 vCredPresCBS Valor do crédito presumido da CBS monofásica. E UB88 N 1 - 1 13v**^

**324.97 UB97 - x- Sequencia XML G UB80 0 - 1 Operações com diferimento, aplicado aos biocombustíveis. Exemplo: operação do produtor de biocombustível (usina).**

**324.98 UB98 pDifIBS Percentual do diferimento do imposto monofásico. E UB94 N 1 - 1** (^) **3v2- 4 A ser aplicado em vIBSMono.
324.99 UB99 vIBSMonoDif Valor do IBS mono diferido. E UB94 N 1 - 1 13v2** (^) **A ser deduzido do valor do IBS.
324.
0 UB100**^ **pDifCBS**^ **Percentual do diferimento do imposto monofásico**^ **E**^ **UB94**^ **N**^^1 **-**^1^ **3v2- 4 A ser aplicado em vCBSMono**^
**324.
101 UB101**^ **vCBSMonoDif**^ **Valor do CBS Mono diferido.**^ **E**^ **UB94**^ **N**^^1 **-**^1^ **13v2**^ **A ser deduzido do valor da CBS.**^
**324.
102 UB102**^ **vTotIBSMono**^ **Total de IBS Monofásico.**^ **E**^ **UB80**^ **N**^^1 **-**^1^ **13v**^
**Considerando a adição do valor de retenção e deduzindo a
parcela diferida conforme o caso.
324.
103 UB103**^ **vTotCBSMono**^ **Total da CBS Monofásica.**^ **E**^ **UB80**^ **N**^^1 **-**^1^ **13v**^
**Considerando a adição do valor de retenção e deduzindo a
parcela diferida conforme o caso.**


```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
5.2 Grupo W. Total da NF-e

```
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
```
```
355.1 W31 IBSCBSSelTot Totais da NF-e com IBS, CBS e IS G A01 0 - 1
```
```
O grupo de valores totais da NF-e deve ser informado com o
somatório do campo correspondente dos itens.
O IBS, a CBS e o IS são por fora, por isso seus valores devem
ser adicionados ao valor total da NF.
```
```
355.2 W32 gSel Grupo total do imposto seletivo G W31 0 - 1
```
```
355.3 W33 vBCSel Total da base de cálculo do imposto seletivo E W32 N 1 - 1 11v0- 4
```
```
355.4 W34 vImpSel Total do imposto seletivo E W32 N 1 - 1 13v2
```
```
355.5 W35 vBCIBSCBS Valor total da BC do IBS e da CBS E W31 N 1 - 1 13v2
```
```
355.6 W36 gIBS Grupo total do IBS G W31 1 - 1
```
```
355.7 W37 gIBSUFTot Grupo total do IBS da UF G W36 1 - 1
```
```
355.8 W38 vCresPres Valor total do crédito presumido E W37 N 1 - 1 13v2
```
```
355.9 W39 vDif Valor total do diferimento E W37 N 1 - 1 13v2
```
```
355.1 W40 vDevTrib Valor total de devolução de tributos E W37 N 1 - 1 13v2
```
```
355.11 W41 vDeson Valor total de desoneração E W37 N 1 - 1 13v2
```
```
355.12 W42 vIBSUF Valor total do IBS da UF E W37 N 1 - 1 13v2
```
```
355.13 W43 gIBSMunTot Grupo total do IBS do Município G W36 1 - 1
```
```
355.14 W44 vCresPres Valor total do crédito presumido E W43 N 1 - 1 13v2
```
```
355.15 W45 vDif Valor total do diferimento E W43 N 1 - 1 13v2
```
```
355.16 W46 vDevTrib Valor total de devolução de tributos E W43 N 1 - 1 13v2
```
```
355.17 W47 vDeson Valor total de desoneração E W43 N 1 - 1 13v2
```
```
355.18 W48 vIBSMun Valor total do IBS do Município E W43 N 1 - 1 13v2
```
```
355.19 W49 vIBSTot Valor total do IBS E W43 N 1 - 1 13v2
```

```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
# ID Campo Descrição Ele Pai Tipo Ocor. Tam. Observação
```
**355.2 W50 gCBS Grupo total da CBS G W36 1 - 1**

**355.21 W51 vCresPres Valor total do crédito presumido E W50 N 1 - 1 13v2**

**355.22 W52 vDif Valor total do diferimento E W50 N 1 - 1 13v2**

**355.23 W53 vDevTrib Valor total de devolução de tributos E W50 N 1 - 1 13v2**

**355.24 W54 vDeson Valor total de desoneração E W50 N 1 - 1 13v2**

**355.25 W55 vCBS Valor total da CBS E W50 N 1 - 1 13v2**

**355.26 W56 gMono Grupo total da Monofasia G W36 0 - 1**

**355.27 W57 vTotIBSMono Total do IBS monofásico E W56 N 1 - 1 13v2**

**355.28 W58 vTotCBSMono Total da CBS monofásica E W56 N 1 - 1 13v2**

**355.29 W59 vTotNF Valor total da NF-e com IBS / CBS / IS E W56 N 1 - 1 13v2 O IBS, a CBS e o IS são por fora, por isso ser adicionados ao valor total da NF. seus valores devem**


```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
6 Regras de Validação

Grupo B. Identificação da Nota Fiscal eletrônica

```
Campo-Seq Modelo Regra de Validação Aplic. Msg Efeito Descrição Erro
```
#### B25- 80 55/65

```
Se finalidade da NF-e igual a crédito ou débito
(tag:finNFe=5 ou 6):
Não pode ser informado ICMS (tag: ICMS), ISSQN (tag:
ISSQN), IPI (tag: IPI), II (tag: II), PIS (tag: PIS), PIS ST (tag:
PISST), COFINS(tag: COFINS), COFINS ST (tag: COFINSST),
ICMS UF Destino (tag: ICMSUFDest) e Imposto Devolvido
(tag: impostoDevol).
```
```
Obrig. 391 Rej.
```
```
Rejeição: NF-e com finalidade de débito ou crédito
somente para IBS/CBS.
```
#### B25- 90 55/65

```
Se finalidade da NF-e diferente de crédito ou débito
(tag:finNFe=5 ou 6):
Deve ser informado ICMS (tag: ICMS) ou ISSQN (tag:
ISSQN).
```
```
Obrig. 392 Rej. Rejeição: NF-e sem informação de ICMS / ISSQN.
```
```
B25b- 50 65
```
```
NF-e com indicativo de NFC-e com entrega a domicílio
(tag:indPres=4) e não informado endereço do
destinatário (id: E5, grupo: enderDest)
```
```
Obrig. 393 Rej.
```
```
Rejeição: NFC-e de entrega a domicílio e não
informado endereço do destinatário.
```

```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
6.1 Grupo UB. Informações dos tributos IBS / CBS e Imposto Seletivo

```
Campo-Seq Modelo Regra de Validação Aplic. Msg Efeito Descrição Erro
```
#### UB02- 10 55/65

```
Não é permitido uso do Imposto Seletivo (grupo: seletivo)
para este cClass.
```
```
Obrig. 363 Rej.
```
```
Rejeição: Não é permitido uso do Imposto Seletivo
para esta classificação da operação.
```
#### UB02- 20 55/65

```
É exigido uso do Imposto Seletivo (grupo: seletivo) para
este cClass.
```
```
Obrig. 367 Rej.
```
```
Rejeição: É exigido o uso do Imposto Seletivo para
esta classificação da operação.
```
#### UB02- 30 55/65

```
É exigido uso do Imposto Seletivo (grupo: seletivo) para
este NCM. Obrig. 368 Rej.
```
```
Rejeição: É exigido o uso do Imposto Seletivo para
esta classificação da operação para este NCM.
```
#### UB03- 10 55/65

```
Se CST do Imposto Seletivo for informado, este deve existir
na tabela de Código de Situação Tributária (tag:
seletivo/CST).
```
```
Obrig. 369 Rej. Rejeição: CST do Imposto Seletivo informado
inexistente
```
#### UB04- 10 55/65

```
Se cClassTrib for informado, este deve existir na tabela de
Classificação Tributária do Imposto Seletivo (tag:
seletivo/cClassTrib)
```
```
Obrig. 370 Rej.
```
```
Rejeição: Classificação Tributária do Imposto Seletivo
informada inexistente
```
#### UB07- 10 55/65

```
Se Se CClassTrib informado exigir grupo do Imposto
Seletivo e Exigir o pImpSel diferente de Zero e pImpSel do
imposto for zero. (tag: seletivo/pSel).
```
```
Obrig. 371 Rej.
```
```
Rejeição: CST/CCLassTrib do Imposto Seletivo obriga
informação de alíquota de Imposto Seletivo
```
#### UB08- 10 55 - 65

```
Se Se CClassTrib informado exigir grupo do Imposto
Seletivo e Exigir o pImpSel diferente de Zero e NCM do
item for 2401, 2402, 2403, 2404, 2203, 2204, 2205, 2206,
2208 e pImpSelEspec.
```
```
Obrig. 372 Rej.
```
```
Rejeição: Obrigatório informação de alíquota
específica de Imposto Seletivo
```
#### UB09- 10 55/65

```
Se CClassTrib informado exigir grupo do Imposto Seletivo
(grupo: seletivo) e unidade tributável (tag: seletivo/uTrib)
e quantidade tributável (tag: seletivo/qtrib) não
informadas ou iguais a zero.
```
```
Obrig. 373 Rej.
```
```
Rejeição: UTrib e QTrib do imposto seletivo não
informados
```

```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
**Campo-Seq Modelo Regra de Validação Aplic. Msg Efeito Descrição Erro**

#### UB102- 10 55/65

```
Se informado grupo do IBS e CBS monofásico (grupo:
gIBSCBSMono):
O valor total do IBS Monofásico (tag: vTotIBSMono)
deverá ser resultante de:
vIBSMono + vIBSMonoReten -vIBSMonoDif
```
```
Observação : Aceitar uma tolerância de 0,01 a mais ou a
menos
```
```
Obrig. 374 Rej.
```
```
Rejeição: Valor do IBS monofásico calculado
incorretamente.
```
#### UB103- 10 55/65

```
Se informado grupo do IBS e CBS monofásico (grupo:
gIBSCBSMono):
O valor total da CBS Monofásica (tag: vTotCBSMono)
deverá ser resultante de:
vCBSMono + vCBSMonoReten -vCBSMonoDif
```
```
Observação : Aceitar uma tolerância de 0,01 a mais ou a
menos
```
```
Obrig. 375 Rej.
```
```
Rejeição: Valor da CBS monofásico calculado
incorretamente.
```
#### UB11- 10 55/65

```
Se informado imposto seletivo (tag: imposto/seletivo):
```
```
Valor do IS (vImpSel) = BC (tag: seletivo/vBC) * Alq (tag:
seletivo/pImpSel)
```
```
Observação 1:
Se informada alíquota específica (tag:
seletivo/pImpSelEspec):
Valor do IS (vImpSel) = qtd (tag: seletivo/vTrib) * Alq (tag:
seletivo/pImpSelEspec)
```
```
Observação 2: Aceitar uma tolerância de 0,01 a mais ou a
menos.
```
```
Obrig. 376 Rej.
```
```
Rejeição: Valor do Imposto Seletivo diferente de Base
de Cálculo x Alíquota
```
#### UB12- 10 55/65

```
Se CST do IBS/CBS for informado, este deve existir na
tabela de Código de Situação Tributária (tag: gIBSCBS/CST) Obrig.^310 Rej.^ Rejeição: CST do IBS/CBS informado inexistente^
```

```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
**Campo-Seq Modelo Regra de Validação Aplic. Msg Efeito Descrição Erro**

#### UB12- 20 55/65

```
Se o CST do IBS/CBS (tag: gIBSCBS/CST) informado VEDAR
preenchimento do grupo de informações específicas do
IBS/CBS, este grupo NÂO DEVE estar informado (grupo:
imposto/IBSCBSSel)
```
```
Obrig. 313 Rej. Rejeição: Grupo IBS/CBS não deve ser preenchido
para o CST informado
```
#### UB12- 30 55/65

```
Se o CST do IBS/CBS (tag: gIBSCBS/CST) informado EXIGIR
preenchimento do grupo de informações específicas do
IBS/CBS, este grupo DEVE estar informado (grupo:
imposto/IBSCBSSel)
```
```
Obrig. 314 Rej.
```
```
Rejeição: Grupo IBS/CBS deve ser preenchido para o
CST informado
```
#### UB13- 10 55/65

```
Se cClassTrib for informado, este deve existir na tabela de
Classificação Tributária do IBS/CBS (tag:
gIBSCBS/cClassTrib)
```
```
Obrig. 311 Rej.
```
```
Rejeição: Classificação Tributária do IBS/CBS
informada inexistente
```
```
UB13- 20 55/65 cClassTrib (tag: gIBSCBS/cClassTrib) for informado deve
ser compatível com CST (tag: gIBSCBS/CST)
```
```
Obrig. 312 Rej. Rejeição: Rejeição: Classificação Tributária
incompatível com o CST informado
```
#### UB17- 10 55/65

```
Alíiquota do IBS da UF (tag: pIBSUF) deve ser igual a 0,1%
para documento com data de emissão no ano de 2026. Art
342 PL 68/24.
```
```
Obrig. 377 Rej.
```
```
Rejeição: Alíiquota do IBS da UF deve ser igual a 0,1%
para documento emitido em 2026
```
#### UB17- 20 55/65

```
Alíiquota do IBS da UF (tag: pIBSUF) deve ser igual a 0,05%
para documento com data de emissão nos anos de 2027 e
```
2028. Art 343 PL 68/24.

```
Obrig. 378 Rej.
```
```
Rejeição: Alíiquota do IBS da UF deve ser igual a 0,05%
para documento emitido em 2027 e 2028
```
#### UB19- 10 55/65

```
Se informado grupo IBS de competência das Unidades
Federadas (gIBSUF)
O valor bruto do tributo do IBS das UF deve ser igual à
multiplicação da Base de Cálculo pela Alíquota.
```
```
vTribOP (tag: gIBSUF/vTribOP) = vBC (tag: gIBSCBS/vBC) *
pIBSUF (tag: gIBSUF/pIBSUF)
```
```
Observação : Aceitar uma tolerância de 0,01 a mais ou a
menos.
```
```
Obrig. 316 Rej.
```
```
Rejeição: NF-e com valor bruto do tributo do IBS das
UFs calculado incorretamente.
```

```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
**Campo-Seq Modelo Regra de Validação Aplic. Msg Efeito Descrição Erro**

#### UB20- 10 55/65

```
Não é permitido o uso de crédito presumido (grupo:
gCredPres) para este cClass. Obrig^379 Rej.^
```
```
Rejeição: Não é permitido o uso de Crédito Presumido
para esta classificação da operação.
```
#### UB20- 20 55/65

```
É exigido o uso de crédito presumido da UF (grupo:
gCredPres) para este cClass. Obrig.^317 Rej.^
```
```
Rejeição: É exigido o uso de Crédito Presumido da UF
para esta classificação da operação.
```
#### UB22- 10 55/65

```
Se informado grupo do crédito presumido (grupo:
gIBSUF/gCredPres):
O valor do Crédito Presumido (tag: vCredPres) deverá ser
resultante da Base de Cálculo x Percentual do Crédito
Presumido (vBC x pCredPres)
```
```
Observação : Aceitar uma tolerância de 0,01 a mais ou a
menos
```
```
Obrig. 318 Rej.
```
```
Rejeição: Valor do Crédito Presumido da UF diferente
de Base de Cálculo x Percentual
```
#### UB23- 10 55/65

```
Não é permitido o uso de diferimento (grupo: gDif) para
este cClass.
```
```
Obrig 380 Rej.
```
```
Rejeição: Não é permitido o uso de Diferimento para
esta classificação da operação.
```
#### UB23- 20 55/65

```
É exigido o uso de diferimento da UF (grupo: gDif) para
este cClass.
```
```
Obrig 319 Rej.
```
```
Rejeição: É exigido o uso de Diferimento da UF para
esta classificação da operação.
```
#### UB25- 10 55/65

```
Se informado grupo do Diferimento (gIBSUF/gDif):
O valor do Diferimento (vDif) deverá ser resultante da
Base de Cálculo x Percentual do Diferimento (vBC x pDif)
```
```
Observação : Aceitar uma tolerância de 0,01 a mais ou a
menos
```
```
Obrig. 320 Rej.
```
```
Rejeição: Valor do Diferimento da UF diferente de
Base de Cálculo x Percentual
```
```
UB28- 10 55/65 Não é permitido o uso de redução de alíquota (grupo:
gRed) para este cClass.
```
```
Obrig 381 Rej. Rejeição: Não é permitido o uso de Redução de
Alíquota para esta classificação da operação.
```
```
UB28- 20 55/65 É exigido o uso de redução de alíquota (grupo: gRed) para
este cClass.
```
```
Obrig 382 Rej. Rejeição: É exigido o uso de Redução de Alíquota para
esta classificação da operação.
```

```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
**Campo-Seq Modelo Regra de Validação Aplic. Msg Efeito Descrição Erro**

#### UB29- 10 55/65

```
Se informado grupo de Redução de Alíquota
(gIBSUF/gRed):
Percentual de Redução de Alíquota (pRedAliq) não é válido
para este cClassTrib (IBSCBSSel/cClassTrib)
```
```
Obrig. 383 Rej. Rejeição: Percentual de redução de alíquota da UF
não é válido para este cClassTrib
```
#### UB30- 10 55/65

```
Se informado grupo de Redução de Alíquota
(gIBSUF/gRed):
Alíquota Efetiva (tag: pAliqEfet) deve ser o resultado da
aplicação do percentual de redução da alíquota (tag:
pRedAliq) na alíquota do IBS da UF (tag: pIBSUF).
Exemplo:
Redução de 40% na alíquota:
Alíquota vigente (A): 10%
Redução na alíquota (R): 40%
Alíquota Efetiva (E): E = A * (1 - R)
E = 10 * (1 - 0,4) = 6
```
```
Obrig. 384 Rej. Rejeição: Valor da Alíquota Efetiva do IBS da UF
calculado incorretamente
```
#### UB31- 10 55/65

```
Não é permitido o uso de desoneração (grupo: gDeson)
para este cClass. Obrig^385 Rej.^
```
```
Rejeição: Não é permitido o uso de Desoneração para
esta classificação da operação.
```
#### UB31- 20 55/65

```
É exigido o uso de desoneração (grupo: gDeson) para este
cClass. Obrig^321 Rej.^
```
```
Rejeição: É exigido o uso de Desoneração da UF para
esta classificação da operação.
```
#### UB32- 10 55/65

```
Se informado grupo da desoneração (gIBSUF/gDeson):
O CST for informado, este deve existir na tabela de
Classificação Tributária do IBS/CBS (tag: gDeson/CST)
```
```
Obrig. 322 Rej.
```
```
Rejeição: CST informado na desoneração da UF
inexistente
```
#### UB33- 10 55/65

```
Se informado grupo da desoneração (gIBSUF/gDeson):
O cClassTrib for informado, este deve existir na tabela de
Classificação Tributária do IBS/CBS (tag:
gDeson/cClassTrib)
```
```
Obrig. 323 Rej.
```
```
Rejeição: Classificação Tributária informada na
desoneração da UF inexistente
```

```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
**Campo-Seq Modelo Regra de Validação Aplic. Msg Efeito Descrição Erro**

#### UB36- 10 55/65

```
Se informado grupo da desoneração (gIBSUF/gDeson):
O valor desonerado (vDeson) deverá ser resultante da
Base de Cálculo x Aliquota (gDeson/vBC x gDeson/pAliq)
```
```
Observação : Aceitar uma tolerância de 0,01 a mais ou a
menos
```
```
Obrig. 324 Rej.
```
```
Rejeição: Valor Desonerado do Município diferente de
Base de Cálculo x Alíquota
```
#### UB37- 10 55/65

```
Se informado grupo IBS de competência das Unidades
Federadas (gIBSUF):
```
```
O valor do IBS (vIBSUF) deverá ser resultante da Base de
Cálculo x Alíquota (vBC [tag: gIBSCBS/vBC] x pIBSUF) -
vCredPres - vDif - vDevTrib
```
```
Observação 1: Aceitar uma tolerância de 0,01 a mais ou a
menos
```
```
Observação 2: Em caso de preenchimento do grupo de
redução (pRed) a alíquota utilizada deverá ser a tag
Alíquota Efetiva (pAliqEfet)
```
```
Observação 3 : Conforme cClass escolhido o valor do
crédito presumido (vCredPres) deve ser subtraído do total
do IBS.
```
```
Obrig. 315 Rej.
```
```
Rejeição: NF-e com valor do IBS da UF calculado
incorretamente.
```
#### UB39- 10 55/65

```
Alíiquota do IBS do município (tag: pIBSMun) deve ser
igual a 0,05% para documento com data de emissão nos
anos de 2027 e 2028. Art 343 PL 68/24.
```
```
Obrig. 386 Rej.
```
```
Rejeição: Alíiquota do IBS da Município deve ser igual
a 0,05% para documento emitido em 2027 e 2028
```

```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
**Campo-Seq Modelo Regra de Validação Aplic. Msg Efeito Descrição Erro**

#### UB41- 10 55/65

```
Se informado grupo IBS de competência dos Municipios
(gIBSMun)
O valor bruto do tributo do IBS dos Municípios deve ser
igual à multiplicação da Base de Cálculo pela Alíquota.
```
```
vTribOP (tag: gIBSMun/vTribOP) = vBC (tag: gIBSCBS/vBC)
* pIBSMun (tag: gIBSMun/pIBSMun)
```
```
Observação : Aceitar uma tolerância de 0,01 a mais ou a
menos.
```
```
Obrig. 326 Rej.
```
```
Rejeição: NF-e com valor bruto do tributo do IBS dos
Municípios calculado incorretamente.
```
#### UB42- 10 55/65

```
É exigido o uso de crédito presumido do Município (grupo:
gCredPres) para este cClass.
```
```
Obrig. 327 Rej.
```
```
Rejeição: É exigido o uso de Crédito Presumido do
Município para esta classificação da operação.
```
#### UB44- 10 55/65

```
Se informado grupo do crédito presumido
(gIBSMun/gCredPres):
O valor do Crédito Presumido (vCredPres) deverá ser
resultante da Base de Cálculo x Percentual do Crédito
Presumido (vBC x pCredPres)
```
```
Observação : Aceitar uma tolerância de 0,01 a mais ou a
menos
```
```
Obrig. 328 Rej.
```
```
Rejeição: Valor do Crédito Presumido dos Municípios
diferente de Base de Cálculo x Percentual
```
```
UB45- 10 55/65 É exigido o uso de diferimento (grupo: gDif) Municipal para
este cClass.
```
```
Obrig 329 Rej. Rejeição: É exigido o uso de Diferimento Municipal
para esta classificação da operação.
```
#### UB47- 10 55/65

```
Se informado grupo do Diferimento (gIBSMun/gDif):
O valor do Diferimento (vDif) deverá ser resultante da
Base de Cálculo x Percentual do Diferimento (vBC x pDif)
```
```
Observação : Aceitar uma tolerância de 0,01 a mais ou a
menos
```
```
Obrig. 330 Rej.
```
```
Rejeição: Valor do Diferimento dos Municípios
diferente de Base de Cálculo x Percentual
```

```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
**Campo-Seq Modelo Regra de Validação Aplic. Msg Efeito Descrição Erro**

#### UB51- 10 55/65

```
Se informado grupo de Redução de Alíquota
(gIBSMun/gRed):
Percentual de Redução de Alíquota (pRedAliq) não é válido
para este cClassTrib (IBSCBSSel/cClassTrib)
```
```
Obrig. 387 Rej. Rejeição: Percentual de redução de alíquota dos
Municípios não é válido para este cClassTrib
```
#### UB52- 10 55/65

```
Se informado grupo de Redução de Alíquota
(gIBSMun/gRed):
O valor da alíquota efetiva (pAliqEfet) deve ser igual a
aplicação da redução da alíquota do IBS dos Municípios
(pRedAliq) na alíquota do IBS dos Municípios (pIBSMun).
Exemplo :
Redução de 40% na alíquota:
Alíquota vigente (A): 10%
Redução na alíquota (R): 40%
Alíquota Efetiva (E): E = A * (1 - R)
E = 10 * (1 - 0,4) = 6
```
```
Obrig. 388 Rej. Rejeição: Valor da Alíquota Efetiva do IBS dos
Municípios calculado incorretamente
```
#### UB53- 10 55/65

```
Código de Classificação Tributária exige informação do
grupo de Desoneração Municipal (gIBSMun/gDeson) Obrig.^331 Rej.^
```
```
Rejeição: CST informado exige informação de
desoneração dos Municípios
```
#### UB54- 10 55/65

```
Se informado grupo da desoneração (gIBSMun/gDeson):
O CST for informado, este deve existir na tabela de
Classificação Tributária do IBS/CBS (tag: gDeson/CST)
```
```
Obrig. 332 Rej.
```
```
Rejeição: CST informado na desoneração dos
Municípios inexistente
```
#### UB55- 10 55/65

```
Se informado grupo da desoneração (gIBSMun/gDeson):
O cClassTrib for informado, este deve existir na tabela de
Classificação Tributária do IBS/CBS (tag:
gDeson/cClassTrib)
```
```
Obrig. 333 Rej.
```
```
Rejeição: Classificação Tributária informada na
desoneração dos Municípios inexistente
```
#### UB58- 10 55/65

```
Se informado grupo da desoneração (gIBSMun/gDeson):
O valor desonerado (vDeson) deverá ser resultante da
Base de Cálculo x Aliquota (gDeson/vBC x gDeson/pAliq)
```
```
Observação : Aceitar uma tolerância de 0,01 a mais ou a
menos
```
```
Obrig. 334 Rej.
```
```
Rejeição: Valor Desonerado do Município diferente de
Base de Cálculo x Alíquota
```

```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
**Campo-Seq Modelo Regra de Validação Aplic. Msg Efeito Descrição Erro**

#### UB59- 10 55/65

```
Se informado grupo IBS de competência dos Municípios
(gIBSMun):
```
```
O valor do IBS (vIBSMun) deverá ser resultante da Base de
Cálculo x Alíquota (vBC [tag: gIBSCBS/vBC] x pIBSMun) -
vCredPres - vDif - vDeson - vDevTrib
```
```
Observação 1: Aceitar uma tolerância de 0,01 a mais ou a
menos
```
```
Observação 2: Em caso de preenchimento do grupo de
redução (pRed) a alíquota utilizada deverá ser a tag
Alíquota Efetiva (pAliqEfet)
```
```
Observação 3: Conforme cClass escolhido o valor do
crédito presumido (vCredPres) deve ser subtraído do total
do IBS.
```
```
Obrig. 325 Rej.
```
```
Rejeição: NF-e com valor do IBS dos Municípios
calculado incorretamente.
```
#### UB63- 10 55/65

```
Se informado grupo CBS (gCBS)
O valor bruto do tributo da CBS deve ser igual à
multiplicação da Base de Cálculo pela Alíquota.
```
```
vTribOP (tag: gCBS/vTribOP) = vBC (tag: gIBSCBS/vBC) *
pCBS (tag: gCBS/pCBS)
```
```
Observação : Aceitar uma tolerância de 0,01 a mais ou a
menos.
```
```
Obrig. 336 Rej. Rejeição: NF-e com valor bruto do tributo da CBS
calculado incorretamente.
```
```
UB64- 10 55/65 É exigido o uso de crédito presumido da CBS (grupo:
gCredPres) para este cClass.
```
```
Obrig. 337 Rej. Rejeição: É exigido o uso de Crédito Presumido da CBS
para esta classificação da operação.
```

```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
**Campo-Seq Modelo Regra de Validação Aplic. Msg Efeito Descrição Erro**

#### UB66- 10 55/65

```
Se informado grupo do crédito presumido
(gCBS/gCredPres):
O valor do Crédito Presumido (vCredPres) deverá ser
resultante da Base de Cálculo x Percentual do Crédito
Presumido (vBC x pCredPres)
```
```
Observação : Aceitar uma tolerância de 0,01 a mais ou a
menos
```
```
Obrig. 338 Rej.
```
```
Rejeição: Valor do Crédito Presumido da CBS
diferente de Base de Cálculo x Percentual
```
#### UB67- 10 55/65

```
É exigido o uso de diferimento (grupo: gDif) da CBS para
este cClass. Obrig^339 Rej.^
```
```
Rejeição: É exigido o uso de Diferimento da CBS para
esta classificação da operação.
```
#### UB69- 10 55/65

```
Se informado grupo do Diferimento (gCBS/gDif):
O valor do Diferimento (vDif) deverá ser resultante da
Base de Cálculo x Percentual do Diferimento (vBC x pDif)
```
```
Observação : Aceitar uma tolerância de 0,01 a mais ou a
menos
```
```
Obrig. 340 Rej.
```
```
Rejeição: Valor do Diferimento da CBS diferente de
Base de Cálculo x Percentual
```
#### UB73- 10 55/65

```
Se informado grupo de Redução de Alíquota (gCBS/gRed):
Percentual de Redução de Alíquota (pRedAliq) não é válido
para este cClassTrib (IBSCBSSel/cClassTrib)
```
```
Obrig. 389 Rej.
```
```
Rejeição: Percentual de redução de alíquota da CBS
não é válido para este cClassTrib
```
#### UB74- 10 55/65

```
Se informado grupo de Redução de Alíquota (gCBS/gRed):
O valor da alíquota efetiva (pAliqEfet) deve ser igual a
aplicação da redução da alíquota da CBS (pRedAliq) na
alíquota da CBS (pCBS).
Exemplo :
Redução de 40% na alíquota:
Alíquota vigente (A): 10%
Redução na alíquota (R): 40%
Alíquota Efetiva (E): E = A * (1 - R)
E = 10 * (1 - 0,4) = 6
```
```
Obrig. 390 Rej.
```
```
Rejeição: Valor da Alíquota Efetiva da CBS calculado
incorretamente
```

```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
**Campo-Seq Modelo Regra de Validação Aplic. Msg Efeito Descrição Erro**

#### UB75- 10 55/65

```
Código de Classificação Tributária exige informação do
grupo de Desoneração da CBS (gCBS/gDeson) Obrig.^341 Rej.^
```
```
Rejeição: CST informado exige informação de
desoneração da CBS
```
#### UB76- 10 55/65

```
Se informado grupo da desoneração (gCBS/gDeson):
O CST for informado, este deve existir na tabela de
Classificação Tributária do IBS/CBS (tag: gDeson/CST)
```
```
Obrig. 342 Rej.
```
```
Rejeição: CST informado na desoneração da CBS
inexistente
```
#### UB77- 10 55/65

```
Se informado grupo da desoneração (gCBS/gDeson):
O cClassTrib for informado, este deve existir na tabela de
Classificação Tributária do IBS/CBS (tag:
gDeson/cClassTrib)
```
```
Obrig. 343 Rej.
```
```
Rejeição: Classificação Tributária informada na
desoneração a CBS inexistente
```
#### UB80- 10 55/65

```
Se informado grupo da desoneração (gCBS/gDeson):
O valor desonerado (vDeson) deverá ser resultante da
Base de Cálculo x Aliquota (gDeson/vBC x gDeson/pAliq)
```
```
Observação : Aceitar uma tolerância de 0,01 a mais ou a
menos
```
```
Obrig. 344 Rej.
```
```
Rejeição: Valor Desonerado da CBS diferente de Base
de Cálculo x Alíquota
```

```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
```
Campo-Seq Modelo Regra de Validação Aplic. Msg Efeito Descrição Erro
```
#### UB81- 10 55/65

```
Se informado grupo CBS (gCBS):
```
```
O valor da CBS (vCBS) deverá ser resultante da Base de
Cálculo x Alíquota (vBC [tag: gIBSCBS/vBC] x pCBS) -
vCredPres - vDif - vDevTrib
```
```
Observação 1: Aceitar uma tolerância de 0,01 a mais ou a
menos
```
```
Observação 2: Em caso de preenchimento do grupo de
redução (pRed) a alíquota utilizada deverá ser a tag
Alíquota Efetiva (pAliqEfet)
```
```
Observação 3: Conforme cClass escolhido o valor do
crédito presumido (vCredPres) deve ser subtraído do total
do IBS.
```
```
Obrig. 335 Rej. Rejeição: NF-e com valor da CBS calculado
incorretamente.
```
Grupo W03. Total da NF-e - IBS / CBS / IS

```
Campo-Seq Modelo Regra de Validação Aplic. Msg Efeito Descrição Erro
```
#### W31- 10 55/65

```
O grupo de totais do IBS, CBS e IS (IBSCBSSelTot) só deve
ser informado se existir pelo menos uma ocorrência de
IBS, CBS ou Imposto Seletivo nos itens.
```
```
Obrig 345 Rej.
```
```
Rejeição: Total de IBS, CBS e IS só deve ser informado
se existir IBS/CBS declarado nos itens do DFe
```
#### W33- 10 55/65

```
O total da BC do imposto seletivo deverá ser a soma dos
campos vBC (tag: seletivo/vBC) informados nos itens.
```
```
Obrig 394 Rej.
```
```
Rejeição: Total da BC do Imposto Seletivo difere da
soma dos itens
```
#### W34- 10 55/65

```
O total do Imposto Seletivo deverá ser a soma dos
campos vImpSel (tag: seletivo/vImpSel) informados nos
itens.
```
```
Obrig 395 Rej.
```
```
Rejeição: Total do Imposto Seletivo difere da soma
dos itens
```

```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
**Campo-Seq Modelo Regra de Validação Aplic. Msg Efeito Descrição Erro**

#### W35- 10 55/65

```
O total da BC do IBS e da CBS de deverá ser a soma dos
campos vBC (tag: gIBSCBS/vBC) informados nos itens
```
```
Obrig. 396 Rej.
```
```
Rejeição: Total da BC do IBS e da CBS difere da soma
dos itens
```
#### W38- 10 55/65

```
O total do Crédito Presumido do IBS UF deverá ser a
soma do campo vCredPres do IBS UF informados nos
itens
```
```
Obrig 346 Rej.
```
```
Rejeição: Total de Crédito Presumido do IBS UF
difere da soma dos itens
```
#### W39- 10 55/65

```
O total do Diferimento do IBS UF deverá ser a soma do
campo vDif do IBS UF informados nos itens
```
```
Obrig 347 Rej.
```
```
Rejeição: Total de Diferimento do IBS UF difere da
soma dos itens
```
#### W40- 10 55/65

```
O total Devolvido do IBS UF deverá ser a soma do campo
vDevTrib do IBS UF informados nos itens
```
```
Obrig 348 Rej.
```
```
Rejeição: Total Devolvido do IBS UF difere da soma
dos itens
```
#### W41- 10 55/65

```
O total Desonerado do IBS UF deverá ser a soma do
campo vDeson do IBS UF informados nos itens
```
```
Obrig 349 Rej.
```
```
Rejeição: Total Desonerado do IBS UF difere da soma
dos itens
```
#### W42- 10 55/65

```
O total do IBS UF deverá ser a soma do campo vIBSUF
informados nos itens
```
```
Obrig 350 Rej. Rejeição: Total de IBS UF difere da soma dos itens
```
#### W44- 10 55/65

```
O total do Crédito Presumido do IBS Municipal deverá ser
a soma do campo vCredPres do IBS Municipal informados
nos itens
```
```
Obrig 351 Rej.
```
```
Rejeição: Total de Crédito Presumido do IBS
Municipal difere da soma dos itens
```
#### W45- 10 55/65

```
O total do Diferimento do IBS Municipal deverá ser a
soma do campo vDif do IBS Municipal informados nos
itens
```
```
Obrig 352 Rej.
```
```
Rejeição: Total de Diferimento do IBS Municipal
difere da soma dos itens
```
#### W46- 10 55/65

```
O total Devolvido do IBS Municipal deverá ser a soma do
campo vDevTrib do IBS Municipal informados nos itens
```
```
Obrig 353 Rej.
```
```
Rejeição: Total Devolvido do IBS Municipal difere da
soma dos itens
```
#### W47- 10 55/65

```
O total Desonerado do IBS Municipal deverá ser a soma
do campo vDeson do IBS Municipal informados nos itens
```
```
Obrig 355 Rej.
```
```
Rejeição: Total Desonerado do IBS Municipal difere
da soma dos itens
```
#### W48- 10 55/65

```
O total do IBS Municipal deverá ser a soma do campo
vIBSMun informados nos itens
```
```
Obrig 354 Rej.
```
```
Rejeição: Total de IBS Municipal difere da soma dos
itens
```

```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
**Campo-Seq Modelo Regra de Validação Aplic. Msg Efeito Descrição Erro**

#### W49- 10 55/65

```
O Total do IBS deverá ser a soma do IBS das UF e do IBS
Municipal (vIBSUF + vIBSMun)
```
```
Obrig 356 Rej.
```
```
Rejeição: Total do IBS difere da soma do IBS UF e IBS
Municipal
```
#### W51- 10 55/65

```
O total do Crédito Presumido do CBS deverá ser a soma
do campo vCredPres do CBS informados nos itens
```
```
Obrig 357 Rej.
```
```
Rejeição: Total de Crédito Presumido do CBS difere
da soma dos itens
```
#### W52- 10 55/65

```
O total do Diferimento do CBS deverá ser a soma do
campo vDif do CBS informados nos itens
```
```
Obrig 358 Rej.
```
```
Rejeição: Total de Diferimento do CBS difere da soma
dos itens
```
#### W53- 10 55/65

```
O total Devolvido do CBS deverá ser a soma do campo
vDevTrib do CBS informados nos itens
```
```
Obrig 359 Rej.
```
```
Rejeição: Total Devolvido do CBS difere da soma dos
itens
```
#### W54- 10 55/65

```
O total Desonerado do CBS deverá ser a soma do campo
vDeson do CBS informados nos itens
```
```
Obrig 360 Rej.
```
```
Rejeição: Total Desonerado do CBS difere da soma
dos itens
```
#### W55- 10 55/65

```
O total do CBS deverá ser a soma do campo vCBS
informados nos itens
```
```
Obrig 361 Rej. Rejeição: Total de CBS difere da soma dos itens
```
#### W57- 10 55/65

```
O total do IBS monofásico deverá ser a soma dos campos
vTOTIBSMono informados nos itens
```
```
Obrig 397 Rej.
```
```
Rejeição: Total do IBS monofásico difere da soma dos
itens
```
#### W58- 10 55/65

```
O total da CBS monofásica deverá ser a soma dos campos
vTOTCBSMono informados nos itens
```
```
Obrig 398 Rej.
```
```
Rejeição: Total da CBS monofásica difere da soma
dos itens
```
#### W59- 10 55/65

```
O total geral do DFe deverá ser a soma do total NF (grupo
total/vNF) + vTotIBS (IBSCBSSelTot/gIBS/vTotIBS) + vCBS
(IBSCBSSelTot/gCBS/vCBS)+ vSel
(IBSCBSSelTot/gSel/vImpSel) + vTotIBSMono
(gIBSCBSMono/vTotIBSMono) +
(gIBSCBSMono/vTotCBSMono) vTotCBSMono
```
```
Obrig 362 Rej.
```
```
Rejeição: Total da NFe difere da soma do total da
Nota Fiscal, IBS, CBS e IS
```

```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
7 Eventos

7.1 **Evento: “Decurso de** Prazo de Internalização na Área de Livre

Comércio - ALC ou Zona Franca de Manaus - **ZFM”**

Função: Evento para marcar as notas não internalizadas na Suframa (evento de internalização) no
prazo regulamentar e que tem por propósito disparar o lançamento do IBS desonerado para o fornecedor
e o estorno do crédito presumido para o destinatário.
Autor do Evento: Fisco das regiões incentivadas (ou Comitê Gestor). -- A definir
Modelo: NF-e modelo 55
Validação: Não deve existir evento Internamento Suframa atrelado à NF-e. Somente pode ser gerado
a partir do decurso do prazo regulamentar.

Código do Tipo de Evento: XXXXXXX

7.2 **Evento: “Cancelamento do** Evento Decurso de Prazo de Internalização

na Área de Livre Comércio - ALC ou Zona Franca de Manaus - **ZFM”**

Função: Cancelar o Evento “Decurso de Prazo de Internalização na Área de Livre Comércio - ALC ou
Zona Franca de Manaus – ZFM”
Modelo: NF-e modelo 55
Validação: Deve existir o evento de “Decurso de prazo de internalização na Área de Livre Comércio -
ALC ou Zona Franca de Manaus – ZFM” identificado pelo respectivo protocolo de autorização
Código do Tipo de Evento: XXXXXXX

7.3 **Evento: “Solicitação de Apropriação de** Crédito Presumido **”**

Função: Evento a ser gerado pelo adquirente em relação às notas fiscais de aquisição de emissão de
terceiros, e que lhe gerem o direito à apropriação de crédito presumido.
Autor: Adquirente/Destinatário (quando os dois estiverem preenchidos, devem ser iguais) da nota fiscal
Exemplo: Emissão de NF-e destinada à ZFM onde o contribuinte emitente não declara, ou declara a
menor o crédito presumido que o destinatário tem direito. Nesta ocasião o destinatário efetua o registro
deste evento. Este evento substitui totalmente os valores de crédito presumido declarados na NF-e
original.
Modelo: NF-e modelo 55
Código do Tipo de Evento: XXXXXXX
Campos:

1. Tipos de Crédito
    a. Aquisição de material para reciclagem de Cooperativa
    b. Aquisição de material intermediário produzido na ZFM (validação que só permite a
       emissão quando o emitente do evento é o destinatário da NF e emitente e destinatário
       estejam em situação habilitada na ZFM).
    c. Aquisição para ALC ou ZFM de região não incentivada (validação que só permite a
       emissão quando o emitente do evento é o destinatário da NF e quando houver evento de
       internamento na Suframa vinculado à respectiva NF-e).
Obs: a identificação do tipo de crédito é importante porque possuem prazos prescricionais
diferentes, gerando controles no sistema de apuração
2. Valor do crédito
Validação:


```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
1. Quando o campo tipo de crédito for igual a 2: emissão do evento permitida somente quando o
    emitente for o destinatário da NF e emitente e destinatário da NF-e estejam na ZFM;
2. Quando o campo tipo de crédito for igual a 3:
    a. emissão do evento permitida somente quando o emitente for o destinatário da NF e
       quando houver evento internamento Suframa vinculado à respectiva NFe.
    b. O valor do crédito é limitado a 7,5% ou 13,5% do valor da operação, dependendo do
       Estado de origem do fornecedor.
    c. Para definir: Prazo para emitir o evento e o termo inicial de contagem desse prazo

7.4 **Evento: “Cancelamento do Evento Solicitação de Apropriação de**

**Crédito Presumido”**

Função: Cancelar o evento “Solicitação de Apropriação de Crédito Presumido”
Modelo: NF-e modelo 55
Validação: Deve existir o evento de “Solicitação de Apropriação de Crédito Presumido” identificado pelo
respectivo protocolo de autorização
Código do Tipo de Evento: XXXXXXX

7.5 **Evento: “Destinação de** Item para Consumo Pessoal **”**

Função: Permitir ao adquirente informar quando uma aquisição for destinada para o consumo de
pessoa física, hipótese em que não haverá direito à apropriação de crédito.
Modelo: NF-e modelo 55
Autor do Evento: Destinatário da NF-e
Código do Tipo de Evento: XXXXXXX
Campos:
● Indicar item
● indicar quantidade

7.6 **Evento: “Cancelamento do Evento** Destinação de Item para Consumo

Pessoal **”**

Função: Cancelar o evento “Destinação de Item para Consumo Pessoal”
Modelo: NF-e modelo 55
Validação: Deve existir o evento de “Destinação de Item para Consumo Pessoal” identificado pelo
respectivo protocolo de autorização
Código do Tipo de Evento: XXXXXXX

7.7 **Evento: “Imobilização de Item”**

Função: Evento a ser gerado pelo adquirente de bem, quando este for integrado ao seu ativo
imobilizado, a fim de viabilizar a adequada identificação, pelos sistemas da administração tributária, de
prazo-limite para apreciação de eventuais pedidos de ressarcimento, nos termos do art. 59, I do PLP
68/2024.
Modelo: NF-e modelo 55
Autor do Evento: Destinatário da NF-e (Adquirente)
Código do Tipo de Evento: XXXXXXX
Campos:
● Indicar item
● indicar quantidade


```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
7.8 **Evento: “Cancelamento do Evento** Imobilização de Item **”**

Função: Cancelar o evento “Imobilização de Item”
Modelo: NF-e modelo 55
Validação: Deve existir o evento de “Imobilização de Item” identificado pelo respectivo protocolo de
autorização
Código do Tipo de Evento: XXXXXXX

7.9 **Evento: “Solicitação de Apropriação de Crédito de Combustível”**

Função: Evento a ser gerado pelo adquirente de combustível listado no art. 167 do PLP 68/2024 e que
pertença à cadeia produtiva desses combustíveis, para solicitar a apropriação de crédito referente à
parcela que for consumida.
Modelo: NF-e modelo 55
Autor do Evento: Destinatário da NF-e (Adquirente de combustível que faça parte da cadeia produtiva
de combustíveis)
Código do Tipo de Evento: XXXXXXX
Campos:
● Indicar item
● indicar quantidade
Validação: Verificar se o item indicado é combustível.

7.10 **Evento: “Cancelamento do Evento** Solicitação de Apropriação de

Crédito de Combustível **”**

Função: Cancelar o evento “Solicitação de Apropriação de Crédito de Combustível”
Modelo: NF-e modelo 55
Validação: Deve existir o evento de “Solicitação de Apropriação de Crédito de Combustível” identificado
pelo respectivo protocolo de autorização
Código do Tipo de Evento: XXXXXXX


```
Reforma Tributária do Consumo
NT-RT 2024.002 Versão 1.00
```
8 DANFE

Alterações no DANFE para exibir informações relativas aos novos tributos estão em estudo, e serão
publicadas em uma nova versão desta Nota Técnica.

ANEXO I - NCM DO IMPOSTO SELETIVO

Link para a tabela:

https://docs.google.com/spreadsheets/d/1TnXQPmAgAyvOgSIznmw1oxdVmvUIqkYTMH0Xxbepby8/e
dit?usp=sharing

ANEXO II - CÓDIGO DE CLASSIFICAÇÃO TRIBUTÁRIA DO

IMPOSTO SELETIVO

Tabela a ser publicada.

ANEXO III - CÓDIGO DE CLASSIFICAÇÃO TRIBUTÁRIA

DO IBS E DA CBS

Tabela a ser publicada.
