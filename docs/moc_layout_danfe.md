# Sistema Nota Fiscal Eletrônica

## Manual de Orientação do Contribuinte

## Anexo II – Manual de Especificações Técnicas do DANFE

## e Código de Barras

## Versão 7. 00 – Outubro de 2020


## MOC 7.0 – Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras



MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

## Controle de Versões

**Versão Publicação Descrição**

7.00 Novembro/

```
Criação deste manual como documento anexo do MOC. Corresponde aos
capítulos 6 e 7 do MOC 6.0 e seus Anexos III, IV, V, VI e VIII, que tratam da
especificação técnica do DANFE.
```

MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

## Histórico de Alterações / Cronograma......................................................................................................

```
Versão Histórico de atualizações Implantação
Homologação
```
```
Implantação
Produção
```
**7.0 0**

```
Publicação dos campos obrigatórios da NF-e no DANFE Simplificado –
Etiqueta (NT 2020.004)
Imediato Imediato
```
**7.0 0**

```
Atualização das informações do Quadro do Transportador, Informações do
local de retirada e Informações do local de entrega (NT 2018.005)
```
**25/02/2019 29/04/**

**7.**

```
Separação dos capítulos 6 e 7 do MOC 6.0 e seus Anexos III, IV, V, VI e VIII,
que tratam da especificação técnica do DANFE, para este manual específico.
```

MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

## 1 Introdução

Este documento é parte integrante do Manual de Orientação do Contribuinte (MOC) e por objetivo a

definição do leiaute da NF-e, modelos 55 e 65.

O Manual de Orientação do Contribuinte 7.0 é composto pelos seguintes documentos:

- MOC – Visão Geral
- MOC – Anexo I – Leiaute NF-e/NFC-e
- MOC – Anexo II – Manual de Especificações Técnicas do DANFE e Código de Barras
- MOC – Anexo III – Manual de Contingência NF-e
- MOC – Anexo IV – Manual de Contingência NFC-e

As informações do DANFE NFC-e estão publicadas no Manual de Especificações Técnicas do DANFE

NFC-e e QR Code, disponível no Portal Nacional da NFC-e

Ao longo deste documento o acrônimo NF-e é utilizado para todas as situações que se aplicam

indistintamente a ambos os modelos de NF-e (55 e 65). Sempre que é necessário identificar um

dos dois modelos em particular, a diferenciação é feita pela expressão respectiva: NF-e modelo 55 ou

NFC-e modelo 65.

## 2 Código de Barras

O padrão de código de barras a ser impresso no DANFE é o CODE-128C. Utilize o código de barras:

a) No caso de DANFE impresso para representar uma NF-e emitida em operação normal ou em

```
contingência utilizando o Sistema de Contingência do Ambiente Nacional: apenas um código de
barras com a chave de acesso do arquivo da nota fiscal eletrônica, descrita no item 3.9.1 e;
```
b) No caso de DANFE impresso para representar uma NF-e emitida nas demais hipóteses de

```
contingência: dois códigos de barras; um para representar a chave de acesso do arquivo da nota
fiscal eletrônica, e outro para representar dados da NF-e emitida em contingência, conforme o item
3.9.2.
```
A impressão dos códigos de barras no DANFE tem a finalidade de facilitar e agilizar a captura de dados

para consulta nos portais estaduais e da Receita Federal do Brasil.

Com a chave de acesso é possível realizar a consulta de uma Nota Fiscal Eletrônica e de sua situação,

bem como visualizar a autorização de uso da mesma. Dentre outras finalidades do código, destacam-

se o registro do trânsito de mercadorias nos Postos Fiscais e, a critério de cada unidade federada, a

disponibilização do arquivo da NF-e consultada.

Os dados adicionais contidos no segundo código de barras serão utilizados para auxiliar o registro do

trânsito de mercadorias acobertadas por notas fiscais eletrônicas emitidas em contingência.

O conjunto de caracteres representativos do Código de Barras CODE-128C encontra-se no Anexo

III.01 deste manual. Para a sua impressão será considerada a seguinte estrutura de simbolização:

```
Margem
clara
```
Start C Dados representados DV Stop

```
Margem
clara
```
- Margem Clara: espaço claro que não contém nenhuma marca legível por máquina, localizado à
    esquerda e à direita do código, a fim de evitar interferência na decodificação da simbologia. A
    margem clara é chamada também de "área livre", "zona de silêncio" ou "margem de silêncio".
- Start C: inicia a codificação dos dados CODE-128C de acordo com o conjunto de caracteres. O Start
    C não representa nenhum caractere.
- Dados representados: caracteres representados no código de barras.


MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

- DV: dígito verificador da simbologia.
- Stop: caractere de parada que indica o final do código ao leitor óptico.

O código de barras deverá ser impresso com os padrões próprios residentes das impressoras de não

impacto (laser ou deskjet) e de impacto (matriciais ou de linhas) a fim de respeitarem os padrões dos

referidos códigos:

- A área reservada no DANFE;
- Largura mínima total do código de barras (considerando o código de barras da chave de
    acesso, com 44 posições):

```
o 6 cm para impressoras de Não Impacto (Laser de Jato de Tinta);
o 11,5 cm para impressora de impacto (Matricial e de linha)
```
- Altura mínima da barra: 0,8 cm;
- Largura mínima da barra: 0,02 cm, conforme explicado a seguir:

```
Considerando que para cada símbolo da barra são codificados dois caracteres, então teremos:
Tamanho do campo = 44 (caracteres) / 2 = 22 (símbolos)
Considerando que cada símbolo possui 11 (módulos) * 22 (símbolos) = 242 posições
Margem clara = deve ter no mínimo a dimensão de 10 (módulos) * 2 = 20 posições
Start C = 11 (módulos) = 11 posições
DV = 11 (módulos) = 11 posições
Stop = 13 (módulos) = 13 posições
Tamanho total da simbologia = 242 + 20 + 11 + 11 + 13 = 297 (posições)
Largura mínima de cada módulo da barra = 6 cm / 297 (posições) = 0,02 cm
```
### 2.1. Cálculo do Dígito Verificador do CODE-128C

O dígito verificador é baseado em um cálculo do módulo 103 considerando a soma ponderada dos

valores de cada um dos dígitos na mensagem que está sendo codificada, incluindo o valor do caractere

de início (start).

Exemplo: consideremos que a chave de acesso fosse apenas de oito caracteres e contivesse o

seguinte número: 09758364

```
Chave de acesso START 09 75 83 64
Sequência A 1 2 3 4
Valor do caractere B 105 9 75 83 64
Valor Ponderado (A X B) C 105 9 150 249 256
```
- Na linha valor do caractere foi incluso o valor 105 que corresponde ao valor do caractere
    de início (start) para o padrão Code C.
- Excetuando o caractere de start, os demais valores dos caracteres coincidem com os
    valores da chave de acesso, isto porque estamos utilizando o padrão Code C de codificação
    que é exclusivamente numérico.
- O dígito verificador do código será o resto da divisão da somatória dos valores ponderados
    dividido por 103 (módulo 103).

Assim o dígito verificador será:

- Valor da soma ponderada = (1x105)+(1x9)+(2x75)+(3x83)+(4x64) = 769
- 769/103 = 7 resta 48, assim o DV é 48

### 2.2. Representação Simbólica do Código.............................................................................................

```
START 09 75 83 64 DV = 48 STOP
B S B S B S B S B S B S B S B S B S B S B S B S B S B S B S B S B S B S B S B S B S B
2 1 1 2 3 2 2 2 1 2 1 3 2 4 1 2 1 1 1 1 4 2 1 2 1 1 1 4 2 2 3 1 3 1 2 1 2 3 3 1 1 1 2
```
A sequência de barras está descrita na tabela do Anexo III.01 deste manual.

```
B = barra preta
S = espaço ou barra branca
```
A numeração acima indica quantas vezes a barra deverá ser impressa no símbolo.


MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

## 3 DANFE

O DANFE é um documento auxiliar impresso em papel com os objetivos de:

a) Acompanhar o trânsito de mercadorias;

b) Colher a firma do destinatário/tomador para comprovação de entrega das mercadorias ou prestação

de serviços;

c) Prover a necessidade de representações impressas adicionais previstas expressamente na

legislação; e

d) Auxiliar a escrituração da NF-e pelo destinatário não credenciado como emissor de NF-e.

O DANFE será impresso:

a) Em condições normais, em qualquer tipo de papel, exceto papel jornal; e

b) Em uma única via, salvo quando houver disposição expressa em outro sentido.

O DANFE emitido para representar NF-e cujo uso foi autorizado em ambiente de homologação sempre

deverá conter a frase “SEM VALOR FISCAL” no quadro “Informações Complementares” ou em marca

d’água destacada.

O DANFE emitido para representar NF-e emitida em contingência deverá conter esta informação em

destaque, conforme disposto no Anexo IV do MOC 7.

O “Valor Aproximado dos Tributos” calculado pela empresa, correspondente a totalidade dos tributos

federais, estaduais e municipais, cuja incidência influa na formação do respectivo preço de venda,

opcionalmente poderá aparecer no DANFE no campo de Informações Adicionais do Produto (tag:

infAdProd, id:V01) e/ou no campo de Informações Complementares da NF-e (tag: infCpl, id:Z03).

O “Valor Aproximado dos Tributos”, poderá opcionalmente constar no DANFE em campo próprio,

conforme segue:

- Quadro de Cálculo do Imposto: incluir nova coluna com o “Valor Aproximado dos Tributos”
    (itens 3.8.1 e 3.8.2);
- Quadro Dados dos Produtos / Serviços: incluir nova coluna com o “Valor Aproximado dos
    Tributos” (itens 3.1.7, 3.8.1 e 3.8.2).

### 3.1. Campos do DANFE

Os campos do DANFE deverão representar o conteúdo das respectivas TAG XML da NF-e, quando

conhecidos no momento da solicitação de autorização de uso. Não poderão ser impressas informações

que não constem do arquivo da NF-e.

O conteúdo dos campos poderá ser impresso em mais de uma linha desde que a leitura possa ser

feita de forma clara.

O item 3.8 deste manual traz a sugestão de tamanhos a serem seguidos para cada campo, que

garantem a legibilidade prevista na legislação. Embora os tamanhos descritos no item 3.8 não sejam

obrigatórios, o DANFE deverá ser impresso conforme um dos modelos permitidos (conforme o item

3.6.3) e utilizando-se os tamanhos mínimos de fonte descritos no item 3.7.

O DANFE deverá conter todos os campos previstos no modelo adotado, com exceção dos campos

não obrigatórios do quadro “Dados dos Produtos/Serviços”, conforme disposto no item 3.1.7.

As regras estabelecidas para a impressão dos campos aplicam-se também para a impressão das

folhas adicionais do DANFE.

#### 3.1.1. Chave de Acesso...................................................................................................................

A chave de acesso será impressa em onze blocos de quatro dígitos cada, com a seguinte máscara:

9999 9999 9999 9999 9999 9999 9999 9999 9999 9999 9999


MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

#### 3.1.2. Dados da NF-e

No caso de emissão de NF-e normal ou em contingência SVC-XX, os campos 1 e 2 serão preenchidos

conforme o item 3.9.1;

No caso de emissão de NF-e em contingência FS ou FS-DA, os campos 1 e 2 serão preenchidos

conforme o item 3.9.2. Observando que no Campo 2, o Código de Barras Adicional “Dados da NF-e”

será impresso em nove blocos de quatro dígitos cada, com a seguinte máscara: 9999 9999 9999 9999

99 99 9999 9999 9999 9999;

No caso de emissão de NF-e em contingência EPEC, os campos 1 e 2 serão preenchidos conforme o

item 3.9.3.

#### 3.1.3. Dados do Emitente

Deverá conter a identificação do emitente, composta no mínimo por:

- nome ou razão social;
- endereço completo (logradouro, número, complemento, bairro, município, UF, CEP); e
- telefone.

Opcionalmente poderá conter logotipo, desde que sua inclusão não prejudique a exibição das

informações obrigatórias.

#### 3.1.4. Informações do local de retirada (NT 2018.005)

Caso haja preenchimento do grupo F - Local de retirada, fica possibilitada a exibição de informações

no DANFE em área especifica, conforme sugestão de modelo abaixo:

#### 3.1.5. Informações do local de entrega (NT 2018.005)

Caso haja preenchimento do grupo G - Local de entrega, fica possibilitada a exibição de informações

no DANFE em área especifica, conforme sugestão de modelo abaixo:

#### 3.1.6. Quadro Fatura/Duplicatas.......................................................................................................

Poderá conter linhas divisórias internas separando as informações. Poderão ser acrescidas ao quadro

outras informações relativas ao assunto, além das informações contidas no grupo de Dados de


MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

Cobrança da NF-e, desde que estas informações adicionais também estejam contidas no arquivo da

NF-e.

#### 3.1.7. Quadro Dados dos Produtos / Serviços

As informações adicionais de produto (TAG <infAdProd>) deverão constar impressas no DANFE logo

abaixo do item ao qual se referirem.

As informações relativas ao Fundo de Combate à Pobreza (FCP) devem ser informadas:

- No campo de "Informações Adicionais do Produto, tag: infAdProd", os valores informados por

item nos campos (vBCFCP, pFCP, vFCP, vBCFCPST, pFCPST, vFCPST), quando existirem.

Sempre que o conteúdo de um mesmo item for impresso utilizando-se mais de uma linha do quadro

de “Dados dos Produtos/Serviços”, deverá ser aplicado um destaque divisório que identifique quais

linhas foram utilizadas para cada item, a fim de distinguir com clareza um item do outro. O destaque

divisório pode ser aplicado com o uso de linha (pontilhadas, continuas, ou tracejada), espaçamento

duplo entre linhas, sombreamento ou qualquer outro recurso ou efeito semelhante que resulte no

destaque divisório.

Exemplo de destaque divisório com linha tracejada:

```
Cód. Produto Descrição do Produto/Serviço NCM
123 Camisa Social Masculina Manga Longa
EAN 7 890123456789
```
61099000

```
124 Camisa Social Masculina Manga Curta
EAN 7890123456790
```
61099000

```
125 Camiseta Polo
EAN 7890123456790
```
61099000


MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

Exemplo de destaque divisório com espaço duplo:

```
Cód. Produto Descrição do Produto/Serviço NCM
123 Camisa Social Masculina Manga Longa
EAN 7890123456789
```
61099000

```
124 Camisa Social Masculina Manga Curta
EAN 7890123456790
```
61099000

```
125 Camiseta Polo
EAN 7890123456790
```
61099000

Exemplo de destaque divisório com sombreamento:

```
Cód. Produto Descrição do Produto/Serviço NCM
123 Camisa Social Masculina Manga Longa
EAN 7890123456789
```
61099000

```
124 Camisa Social Masculina Manga Curta
EAN 7890123456790
```
61099000

```
125 Camiseta Polo
EAN 7890123456790
```
61099000

Essa exigência também se aplica no caso da utilização de uma mesma coluna para aposição de outro

campo, conforme o item 3.2.

Deve-se utilizar o quadro “Dados dos Produtos/Serviços” para detalhar as operações que não

caracterizem circulação de mercadorias ou prestações de serviços, e que exijam emissão de

documentos fiscais (como transferência de créditos ou apropriação de incentivos fiscais, por exemplo).

Nas situações em que o valor unitário comercial for diferente do valor unitário tributável, ambas as

informações deverão estar expressas e identificadas no DANFE, podendo ser utilizada uma das linhas

adicionais previstas, ou o campo de informações adicionais.

Independentemente do descrito no item 3.3, o contribuinte poderá suprimir colunas do quadro “Dados

dos Produtos/Serviços” que não se apliquem a suas atividades e acrescentar outras do seu interesse.

A inserção destas colunas será realizada à direita da coluna “Descrição dos Produtos/Serviços”. A

ordem das colunas remanescentes deverá ser respeitada.

As seguintes colunas não poderão ser suprimidas:

- Código dos Produtos/Serviços;
- Descrição dos Produtos/Serviços;
- NCM;
- CST;
- CFOP;
- Unidade;
- Quantidade;
- Valor Unitário;
- Valor Total;
- Base de Cálculo do ICMS próprio;
- Valor do ICMS próprio; e
- Alíquota do ICMS.

#### 3.1.8. Informações Complementares

Deverá conter todas as Informações Adicionais da NF-e incluídas nas TAGs <infAdFisco> e <infCpl>,

ficando facultada a impressão das informações adicionais contidas nas TAGs <obsCont>. Na hipótese


MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

de insuficiência de espaço no quadro de “informações complementares”, a impressão destas deverá

ser continuada no verso ou na folha seguinte, neste mesmo quadro ou no quadro “Dados dos

Produtos/Serviços”.

As empresas remetentes devem informar, no campo de “Informações Complementares”, os valores

descritos no grupo de tributação do ICMS para a UF de destino. (NT 2015.003)

Exemplo 1 de preenchimento do DANFE (1ª situação da sistemática de cálculo descrita a seguir):

Exemplo 2 de preenchimento do DANFE (2ª situação da sistemática de cálculo descrita a seguir):

As informações relativas ao Fundo de Combate à Pobreza (FCP) devem ser informadas:

- Os valores de totais do FCP (id: W04b e W06a) devem ser informados em "Informações

Adicionais de Interesse do Fisco, campo “infAdFisco", quando existirem."

#### 3.1.9. Reservado ao Fisco

O contribuinte não deverá preencher este quadro, sendo seu preenchimento de uso exclusivo do fisco,

exceto, a critério da UF, quanto à orientação de impressão do teor das tags contidas no XML de retorno

de autorização da NF-e. Em caso de utilização de formulário de segurança provido de estampa fiscal,

esse quadro não estará presente."

#### 3.1.10. Quadro do Transportador

O campo identificação da Modalidade do Frete (id: X02, tag:modFrete) deverá ser preenchido com um

dos seguintes códigos (NT 2016/002) (Atualizado NT 2108.005):

```
0=Contratação do Frete por conta do Remetente (CIF);
1=Contratação do Frete por conta do Destinatário (FOB);
2=Contratação do Frete por conta de Terceiros;
3=Transporte Próprio por conta do Remetente;
4=Transporte Próprio por conta do Destinatário;
9=Sem Ocorrência de Transporte.
```
Exemplo de preenchimento:

Nome / Razão Social Frete por Conta Código ANTT

(^0) – Remetente
**INFORMAÇÕES COMPLEMENTARES:**
Valores totais do ICMS Interestadual: DIFAL da UF destino R$216,00 + FCP
R$40,00; DIFAL da UF Origem R$324,00.
**INFORMAÇÕES COMPLEMENTARES:**
Valores totais do ICMS Interestadual: DIFAL da UF destino R$156,00 + FCP
R$40,00; DIFAL da UF Origem R$234,00.


MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

### dos Produtos/Serviços” 3.2. Possibilidade de Uso de Uma Mesma Coluna Com Mais de Um Campo no Quadro “Dados

### no Quadro “Dados dos Produtos/Serviços”

É permitida a utilização de uma mesma coluna para aposição de outro campo no quadro “Dados dos

Produtos/Serviços” do DANFE.

A utilização de uma mesma coluna para mais de um campo implicará na ocupação de duas linhas do

“Dados dos Produtos/Serviços” para cada item da NF-e, além das linhas adicionais previstas para

descrever as informações adicionais de produto/serviço (TAG <infAdProd>).

Deverá ser observada a necessidade de aposição de destaque divisório dos diferentes itens do quadro

“Dados dos Produtos/Serviços”, conforme descrito no item 3.1.7.

Os campos que podem ser colocados na mesma coluna são:

- “Código do Produto/Serviço” com “NCM/SH”;
- “CST” com “CFOP”;
- “CSOSN” com “CFOP”;
- “Quantidade” com “Unidade”;
- “Valor Unitário” com “Desconto”;
- “Valor Total” com “Base de Cálculo do ICMS”;
- “Base de Cálculo do ICMS por Substituição Tributária” com “Valor do ICMS por Substituição
    Tributária”;
- “Valor do ICMS Próprio” com “Valor do IPI”;
- “Alíquota do ICMS” com “Alíquota do IPI”.

A utilização de uma mesma coluna para mais de um campo não se aplicará para a aposição do campo

Descrição dos Produtos e/ou Serviços, podendo-se, neste caso, utilizar mais linhas para aposição de

seu conteúdo.

### 3.3. Supressões e Modificações Permitidas

Além das supressões e inclusões de colunas tratadas no item 3.1. 7 poderão ser feitas ainda as

seguintes alterações:

#### 3.3.1. Bloco de Canhoto

Caso o emitente não utilize o bloco de Canhoto, poderá aumentar o quadro “Dados dos

Produtos/Serviços” suprimindo os campos do referido bloco e deslocando para cima os campos

seguintes. Estes ajustes deverão ser feitos no mesmo valor da redução obtida com a eliminação do

quadro Fatura e de sua descrição.

Para a impressão de DANFE que não utilizar formulário de segurança, o bloco de canhoto poderá ser

deslocado para a extremidade inferior do formulário, sem alterações nas demais dimensões e

disposições de campos e quadros.

Essas alterações serão admitidas somente no formato retrato.

#### 3.3.2. Quadro “Fatura/Duplicatas”

O quadro “fatura/duplicatas” poderá ser suprimido, caso o contribuinte não utilize esses documentos;

ou reduzido, desde que contenha todos os dados das respectivas TAGs.

O valor obtido com a eliminação ou redução do quadro “fatura/duplicatas” deverá ser acrescido na

altura do quadro “Dados dos Produtos/Serviços”, deslocando para cima os campos seguintes ao

quadro Fatura e anteriores ao quadro a ser aumentado.

Essas alterações poderão ser feitas tanto nos formatos retrato quanto paisagem.


MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

#### 3.3.3. Quadro “Cálculo do ISSQN”..................................................................................................

Caso não se aplique às suas operações, o emitente poderá suprimir os campos do bloco “Cálculo do

ISSQN” e efetuar os seguintes ajustes:

- Aumentar a altura do quadro “Dados dos Produtos/Serviços” no mesmo valor da redução obtida

com a eliminação dos campos do referido bloco.

- Aumentar a altura do campo “Informações Complementares” e do quadro “Reservado ao Fisco” no

mesmo valor da redução obtida com a eliminação dos campos do bloco “Cálculo do ISSQN”.

### 3.4. Verso do DANFE

Até 50% do verso de qualquer folha do DANFE poderá ser utilizado para continuação dos dados do

quadro “Dados dos Produtos/Serviços”, do campo “Informações Complementares” ou para uma

combinação de ambos. O restante do verso deverá ser deixado sem nenhum tipo de impressão.

Sempre que o verso do DANFE for utilizado, a informação “CONTINUA NO VERSO” deverá constar

no anverso, ao final dos quadros “Dados dos Produtos/Serviços” e “Informações Complementares”,

conforme a utilização.

### 3.5. Folhas Adicionais

O DANFE poderá ser emitido em mais de uma folha.

Cada uma das folhas adicionais deverá conter, na parte superior, no mínimo as seguintes informações,

impressas na mesma disposição e tamanho definidos para a primeira folha:

- Dados de Identificação do Emitente;
- As descrições “DANFE” em destaque, e “Documento Auxiliar da Nota Fiscal Eletrônica”;
- O número e a série da NF-e, o tipo de operação, se Entrada ou Saída, além do número total de

folhas e o número de ordem de cada folha;

- Código(s) de Barras;
- Campos Natureza da Operação e Chave de Acesso; e
- Demais campos de identificação do Emitente: Inscrição Estadual, Inscrição Estadual do Substituto

Tributário e CNPJ.

A área restante das folhas adicionais poderá ser utilizada exclusivamente para apor:

- Os demais itens da NF-e que não couberem na primeira folha do DANFE, mantendo-se as mesmas

colunas com a mesma disposição e largura utilizadas na primeira folha; e/ou

- As demais informações complementares da NF-e que não couberem no campo próprio da primeira

folha do DANFE.

### 3.6. Formulário.................................................................................................................................

Para a impressão do DANFE poderá ser utilizado qualquer tipo de papel, com exceção de papel jornal,

desde que seja garantido o contraste necessário para assegurar leitura dos códigos de barras sem

problemas.

#### 3.6.1. Tamanho do Papel...............................................................................................................

A impressão do DANFE poderá ser efetuada tanto em modo retrato quanto em modo paisagem,

utilizando-se formulários de tamanho mínimo A-4 e máximo Ofício II (230 x 330 mm).

Em caso de uso de folha de tamanho superior ao tamanho A-4 o espaço excedente deverá ser alocado

da seguinte maneira:

- Na horizontal, para aumentar a largura dos campos; e
- Na vertical, somente para aumentar a altura:

```
o do quadro “Dados dos Produtos/Serviços”; ou
o simultaneamente dos campos “Informações Complementares” e “Reservado ao Fisco”;
ou, ainda,
o de uma combinação destas duas opções.
```

MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

#### 3.6.2. Margem Lateral no Formulário

As Margens entre o corpo impresso do DANFE e o final do formulário (ou a linha de picote) deverão

ter, no mínimo, 0,2 cm e, no máximo, 0,8 cm em cada lateral (inclusive nas margens superior e inferior).

#### 3.6.3. Modelos de DANFE Permitidos.............................................................................................

É opção do contribuinte a utilização em folhas soltas ou formulário contínuo, pré-impresso ou em

branco. Poderão ser utilizados os formatos a seguir, devendo a disposição de campos

obrigatoriamente obedecer ao disposto no respectivo anexo:

- Tamanho A-4 em modo retrato:

```
o Folhas Soltas – Anexo III.
o Formulário Contínuo – Anexo III.
```
- Tamanho A-4 em modo paisagem:

```
o Folhas Soltas – Anexo III.
o Formulário Contínuo – Anexo III.
```
### 3.7. Padrões de Caracteres (Tipos de Fontes)

Todos os caracteres deverão estar impressos na fonte Times New Roman ou na fonte Courier New. A

impressão dos dados variáveis feitas por Impressoras de Impacto (Matricial e de Linha) deverá estar

entre 10 e 17 CPP (Caracteres por Polegada).

#### 3.7.1. Descritivo dos Blocos de Campos

Deverá ter tamanho mínimo de cinco (5) pontos, impresso em negrito em caixa alta (maiúsculas).

#### 3.7.2. Descritivo dos Campos do Quadro “Dados dos Produtos/Serviços”

Deverá ser impresso em caixa alta (maiúsculas), com tamanho mínimo de cinco (5) pontos.

#### 3.7.3. Descritivo dos Demais Campos

Deverá ser impresso em caixa alta (maiúsculas) e ter tamanho mínimo de seis (6) pontos.

#### 3.7.4. Conteúdo do Bloco de Campos de Identificação do Documento..............................................

O conteúdo dos campos “DANFE”, “entrada ou saída”, “número”, “série” e “folhas do documento”

deverá ser impresso em caixa alta (maiúsculas). Além disto:

- a descrição “DANFE” deverá estar impressa em negrito e ter tamanho mínimo de doze (12)

pontos, ou 10 CPP;

- a série e número da NF-e, o número de ordem da folha, o total de folhas do DANFE e o número

```
identificador do tipo de operação (se “ENTRADA” ou “SAÍDA”, conforme tag “tpNF”) deverão estar
impressos em negrito e ter tamanho mínimo de dez (10) pontos, ou 10 CPP;
```
- a identificação “DOCUMENTO AUXILIAR DA NOTA FISCAL ELETRÔNICA” e as descrições do

```
tipo de operação, “ENTRADA” ou “SAÍDA” deverão ter tamanho mínimo de oito (8) pontos, ou 17
CPP.
```
#### 3.7.5. Conteúdo do Campo Chave de Acesso.

Deverá ser impresso em formato negrito.

#### 3.7.6. Conteúdo do Quadro Dados do Emitente

Deverá estar impresso em negrito. A razão social e/ou nome fantasia deverá ter tamanho mínimo de

doze (12) pontos, ou 17 CPP e os demais dados do emitente, endereço, município, CEP, fone/fax

deverão ter tamanho mínimo de oito (8) pontos, ou 17 CPP.


MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

#### 3.7.7. Conteúdo dos Campos do Quadro “Dados dos Produtos/Serviços”

Deverá ter tamanho mínimo de seis (6) pontos, ou 17 CPP.

#### 3.7.8. Conteúdo do Campo Informações Complementares

Deverá ter tamanho mínimo de seis (6) pontos, ou 17 CPP.

#### 3.7.9. Conteúdo dos Demais Campos

Deverá ter tamanho mínimo de dez (10) pontos, ou 17 CPP.

### 3.8. Tamanho dos Campos

Esta seção apresenta a sugestão de tamanho e posição de cada campo. Todas as medidas estão em

centímetros.

#### 3.8.1. Formulário A-4 em Modo Retrato

O eixo 0 (zero) é no início da folha no canto superior esquerdo.

NOME (^) Id
da
TAG
Tamanhos
Mínimos
Posição c/ relação
à margem
Linha
Outras
TAG/
Obs
Tam.
das
TAG

##### BLOCO

```
CAMPO Altura Largura Esquerda Superior
CANHOTO
RECEBEMOS DE... 0,85 16,10 0,25 0,
NF-e / Nº 000.000.000 / SÉRIE 000 1,70 4,50 16,35 0,
DATA DE RECEBIMENTO 0,85 4,10 0,25 1,
IDENTIFICAÇÃO E ASSINATURA... 0,85 12,10 4,35 1,
DADOS DA NF-e
QUADRO IDENTIFICAÇÃO DO EMITENTE
Mat.
Laser
```
##### 3,

##### 3.

##### 5,

##### 10.

##### 0,

##### 0.

##### 2,

##### 2.

```
Obs 5
```
```
QUADRO DA DESCRIÇÃO "DANFE..."
3,
3.
```
##### 2,

##### 2.

##### 5,

##### 10.

##### 2,

##### 2.

##### QUADRO CÓDIGO DE BARRAS DA CHAVE

```
Mat.
Laser
```
##### 1,

##### 1.

##### 12,

##### 8.

##### 8,

##### 12.

##### 2,

##### 2.

##### CÓDIGO DE BARRAS DA CHAVE 1,00 11,50 8,62 2,

##### CHAVE DE ACESSO 0,85 12,70 8,12 4,02 44

```
QUADRO TIPO DE OPERAÇÃO Invisível Obs 6
QUADRO NÚMERO/SÉRIE DA NF-e Invisível Obs 7
QUADRO CÓDIGO DE BARRAS DOS DADOS
Mat.
Laser
```
##### 1,

##### 1.

##### 12,7 0

##### 8.

##### 8,

##### 12.

##### 4,

```
4.98 Obs 9
CÓDIGO DE BARRAS DOS DADOS 1,00 7,00 Ver Ver Obs 9
NATUREZA DA OPERAÇÃO B04 0,85 7,87 0,25 6,46 60
DADOS DA NF-e
Mat.
Laser
```
##### 0,

##### 0.

##### 12,

##### 8.

##### 8,

##### 12.

##### 6,

```
6.46 Obs 9 44
INSCRIÇÃO ESTADUAL DO EMITENTE C17 0,85 6,86 0,25 7,31 14
INSCRIÇÃO ESTADUAL DE ST DO EMITENTE C18 0,85 6,86 7,11 7,31 14
CNPJ DO EMITENTE C02 0,85 6,86 13,97 7,31 14
DESTINATÁRIO/REMETENTE 0,42 3,30 0,25 8,16 Invisível
RAZÃO SOCIAL E04 0,85 12,32 0,25 8, 58 60
CNPJ E02 0,85 5,33 12,57 8,58 Negrito 14
DATA DA EMISSÃO B09 0, 85 2,92 17,90 8,58 10
ENDEREÇO E06 0,85 10,16 0,25 9,43 E07 120
BAIRRO/DISTRITO E09 0,85 4,83 10,41 9,43 60
CEP E13 0,85 2,67 15,24 9,43 8
DATA DA ENTRADA/SAÍDA B10 0,85 2,92 17,91 9,43 Negrito 10
MUNICÍPIO E11 0,85 7,11 0,25 10,28 60
FONE/FAX E16 0,85 4,06 7,36 10,28 10
UF E12 0,85 1,14 11,42 10,28 2
INSCRIÇÃO ESTADUAL E03 0,85 5,33 12,56 10,28 14
HORA DA ENTRADA/SAÍDA 0,85 2, 92 17,89 10,28 Negrito
```
```
FATURA/DUPLICATAS 0,42 1,00 0,25 11,09 Invisível
FATURA Y02 0,85 20, 57 0,25 11,51 Obs 1
CÁLCULO DO IMPOSTO 0,42 5,60 0,25 12,36 Invisível
BASE DE CÁLCULO DO ICMS W03 0,85 4,06 0,25 12,78 15
VALOR DO ICMS W04 0,85 4,06 4,31 12,78 15
BASE DE CÁLCULO DO ICMS ST W05 0,85 4,06 8,37 12,78 15
VALOR DO ICMS ST W06 0,85 4,06 12,43 12,78 15
VALOR TOTAL DOS PRODUTOS W07 0,85 4,32 16,49 12,78 15
VALOR DO FRETE W08 0,85 3,30 0,25 13,63 15
VALOR DO SEGURO W09 0,85 3,30 3,55 13,63 15
DESCONTO W10 0,85 3,30 6,85 13,63 15
OUTRAS DESPESAS ACESSÓRIAS W15 0,85 3,30 10,15 13,63 15
VALOR DO IPI W12 0,85 3,30 13,45 13,63 15
VALOR TOTAL DA NOTA W16 0,85 4,06 16,75 13,63 Negrito 15
```

MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

NOME (^) Id
da
TAG
Tamanhos
Mínimos
Posição c/ relação
à margem
Linha
Outras
TAG/
Obs
Tam.
das
TAG

##### BLOCO

```
CAMPO Altura Largura Esquerda Superior
TRANSPORTADOR/VOLUMES TRANSPORTADOS 0,42 5,20 0,25 14,48 Invisível
RAZÃO SOCIAL X06 0,85 9,02 0,25 14,90 60
FRETE POR CONTA DE 0,85 2,79 9,27 14,90 Obs 8
CÓDIGO ANTT X21 0,85 1,78 12,06 14,90 X25 20
PLACA DO VEÍCULO X19 0,85 2,29 13,84 14,90 X23 8
UF X10 0,85 0,76 16,13 14,90 2
CNPJ/CPF X04 0,85 3,94 16,89 14,90 14
ENDEREÇO X08 0,85 9,02 0,25 15,75 60
MUNICÍPIO X09 0,85 6,86 9,27 15,75 60
UF X10 0,85 0,76 16,13 15,75 2
INSCRIÇÃO ESTADUAL X07 0,85 3,94 16,89 15,75 14
QUANTIDADE DE VOLUMES X27 0,85 2,92 0,25 16,60 15
ESPÉCIE X28 0,85 3,05 3,17 16,60 60
MARCA X29 0,85 3,05 6,22 16,60 60
NUMERAÇÃO X30 0,85 4,83 9,27 16,60 60
PESO BRUTO X32 0,85 3,43 14,10 16,60 15
PESO LÍQUIDO X31 0,85 3,30 17,53 16,60 15
```
```
DADOS DOS PRODUTOS/SERVIÇOS 0,42 4,00 0,25 17,45 Invisível
QUADRO DADOS DOS PRODUTOS/SERVIÇOS 6,77 20,5 7 0,25 17,87 Obs 4
CÓDIGO I02 60
DESCRIÇÃO DOS PRODUTOS/SERVIÇOS I04 120
"COLUNAS ESPECÍFICAS DA EMPRESA" Obs 2
NCM/SH I05 8
CST N11 N
CFOP I08 4
UNIDADE I09 I13 6
QUANTIDADE I10 I14 12
```
```
VALOR UNITÁRIO I10a
I14a 16
DESCONTO I17 15
VALOR TOTAL I11 Obs 3 15
B.CÁLC.ICMS N15 15
B.CÁLC.ICMS ST N21 15
VALOR ICMS N17 15
VALOR ICMS ST N23 15
VALOR IPI O14 15
ALÍQUOTA ICMS N16 5
ALÍQUOTA IPI O13 5
CÁLCULO DO ISSQN 0,42 2,29 0,25 24,64 Invisível
INSCRIÇÃO MUNICIPAL C19 0,85 5,08 0,25 25,06 15
VALOR TOTAL DOS SERVIÇOS W18 0,85 5,08 5,33 25,06 15
BASE DE CÁLCULO DO ISSQN W19 0,85 5, 08 10,41 25,06 U02 15
VALOR DO ISSQN W20 0,85 5,33 15,49 25,06 U04 15
DADOS ADICIONAIS 0,42 2,29 0,25 25,91 Invisível
INFORMAÇÕES COMPLEMENTARES Z02 3,07 12,95 0,25 26,33 Z03 5256
RESERVADO AO FISCO Invisível
RESERVADO AO FISCO 3,07 7,62 13,17 26,
Obs 1: Permite-se a inclusão dos dados de duplicatas das TAG do grupo Y
Obs 2: Detalhamento específ icos de produtos/serviços (outras TAG do grupo H)
Obs 3: Total Bruto (TAG) ou Líquido (Mod.1/1-A)?
Obs 4: Colunas apresentadas na ordem descrita
Obs 5: TAG: C03, C04, C06, C07, C08, C09, C11, C12, C13, C
Obs 6: TAG: B
Obs 7: TAG: B07, B
Obs 8: TAG: X
Obs 9: Campo utilizado exclusivamente no Modelo de Contingência
```
#### 3.8.2. Formulário A-4 em Modo Paisagem

O eixo 0 (zero) é no início da folha no canto superior esquerdo.

NOME (^) Id
da
TAG
Tamanho
Mínimo
Posição c/ relação
à margem
Linha
Outras
tag/
obs
Tam.
das
TAG

##### BLOCO

```
CAMPO Altura Largura Esquerda Superior
CANHOTO
NF-e / Nº 000.000.000 / SÉRIE 000 4,53 2,03 0,13 0,
RECEBEMOS DE... 16,95 1,02 0,13 5,
IDENTIFICAÇÃO E ASSINATURA... 9,21 1,02 1,15 5,
DATA DE RECEBIMENTO 6,75 1,05 1,15 14,
DADOS DA NF-e
QUADRO IDENTIFICAÇÃO DO EMITENTE 3,10 11,43 2,41 0,47 Obs 5
QUADRO DA DESCRIÇÃO "DANFE..." 3,10 3,05 13,84 0,
QUADRO CÓDIGO DE BARRAS DA CHAVE 1,19 12,57 16,89 0,
CÓDIGO DE BARRAS DA CHAVE
CHAVE DE ACESSO 0,64 12,57 16,89 1,66 44
QUADRO TIPO DE OPERAÇÃO Invisível Obs 6
QUADRO CÓDIGO DE BARRAS DOS DADOS 1,19 12,57 16,89 2,38 Obs 9
CÓDIGO DE BARRAS DOS DADOS Obs 9
QUADRO NÚMERO/FL./SÉRIE DA NF-e Invisível Obs 7
```

## MOC 7.0 – Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

NOME (^) Id
da
TAG
Tamanho
Mínimo
Posição c/ relação
à margem
Linha
Outras
tag/
obs
Tam.
das
TAG

##### BLOCO

HORA DA ENTRADA/SAÍDA 0,64 4,32 25,14 6,13 Negrito

- Controle de Versões Sumário
- Histórico de Alterações / Cronograma......................................................................................................
- 1 Introdução
- 2 Código de Barras
   - 2.1. Cálculo do Dígito Verificador do CODE-128C
   - 2.2. Representação Simbólica do Código.............................................................................................
- 3 DANFE
   - 3.1. Campos do DANFE
      - 3.1.1. Chave de Acesso...................................................................................................................
      - 3.1.2. Dados da NF-e
      - 3.1.3. Dados do Emitente
      - 3.1.4. Informações do local de retirada (NT 2018.005)
      - 3.1.5. Informações do local de entrega (NT 2018.005)
      - 3.1.6. Quadro Fatura/Duplicatas.......................................................................................................
      - 3.1.7. Quadro Dados dos Produtos / Serviços
      - 3.1.8. Informações Complementares
      - 3.1.9. Reservado ao Fisco
      - 3.1.10. Quadro do Transportador
   - dos Produtos/Serviços” 3.2. Possibilidade de Uso de Uma Mesma Coluna Com Mais de Um Campo no Quadro “Dados
   - 3.3. Supressões e Modificações Permitidas
      - 3.3.1. Bloco de Canhoto
      - 3.3.2. Quadro “Fatura/Duplicatas”
      - 3.3.3. Quadro “Cálculo do ISSQN”..................................................................................................
   - 3.4. Verso do DANFE
   - 3.5. Folhas Adicionais
   - 3.6. Formulário.................................................................................................................................
      - 3.6.1. Tamanho do Papel...............................................................................................................
      - 3.6.2. Margem Lateral no Formulário
      - 3.6.3. Modelos de DANFE Permitidos.............................................................................................
   - 3.7. Padrões de Caracteres (Tipos de Fontes)
      - 3.7.1. Descritivo dos Blocos de Campos
      - 3.7.2. Descritivo dos Campos do Quadro “Dados dos Produtos/Serviços”
      - 3.7.3. Descritivo dos Demais Campos
      - 3.7.4. Conteúdo do Bloco de Campos de Identificação do Documento..............................................
      - 3.7.5. Conteúdo do Campo Chave de Acesso.
      - 3.7.6. Conteúdo do Quadro Dados do Emitente
      - 3.7.7. Conteúdo dos Campos do Quadro “Dados dos Produtos/Serviços”
      - 3.7.8. Conteúdo do Campo Informações Complementares
      - 3.7.9. Conteúdo dos Demais Campos
   - 3.8. Tamanho dos Campos
      - 3.8.1. Formulário A-4 em Modo Retrato
      - 3.8.2. Formulário A-4 em Modo Paisagem
   - 3.9. Campos de Conteúdo Variável
      - 3.9.1. Emissão Normal da NF-e e SVC-XX
      - 3.9.2. Emissão da NF-e em Contingência com Impressão do DANFE em Formulário de Segurança...
      - 3.9.3. Emissão da NF-e com Prévio Registro do EPEC no Ambiente Nacional
   - 3.10. Outros
      - 3.10.1. Marca d’Água
      - 3.10.2. Impressão do Número da Folha
      - 3.10.3. Limitações da Impressora..................................................................................................... MOC 7.0 – Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras
      - 3.10.4. Código de Barras
      - 3.10.5. Campo “Valor de ICMS Desonerado”
   - 3.11. DANFE Simplificado
      - 3.11.1. Tipo e tamanho do Papel......................................................................................................
      - 3.11.2. Chave de acesso
      - 3.11.3. Padrões de Caracteres (Tipos de Fontes)
      - 3.11.4. Campos obrigatórios
   - 3.12. DANFE Simplificado – Etiqueta (NT 2020.004)
      - 3.12.1. Tipo e tamanho do Papel......................................................................................................
      - 3.12.2. Chave de acesso
      - 3.12.3. Padrões de Caracteres (Tipos de Fontes)
      - 3.12.4. Campos obrigatórios
- Anexo III.01 – Conjunto de Caracteres Código de Barras CODE-128C
- Anexo III.02 – DANFE Tamanho A-4 em Modo Retrato, Folhas Soltas
- Anexo III.03 – DANFE Tamanho A-4 em Modo Retrato, Formulário Contínuo
- Anexo III.04 – DANFE Tamanho A-4 em Modo Paisagem, Folhas Soltas............................................
- Anexo III.05 - DANFE Tamanho A-4 em Modo Paisagem, Formulário Contínuo
   - DADOS DA NF-e 0,64 12,57 16,89 3,57 Obs CAMPO Altura Largura Esquerda Superior
   - NATUREZA DA OPERAÇÃO B04 0,64 13,97 2,92 3,57
   - INSCRIÇÃO ESTADUAL DO EMITENTE C17 0,64 8,89 2,92 4,21
   - INSCRIÇÃO ESTADUAL DE ST DO EMITENTE C18 0,64 8,89 11,81 4,21
   - CNPJ DO EMITENTE C02 0,64 8,76 20,70 4,21
   - DESTINATÁRIO/REMETENTE 1,92 0,51 2,41 4,
   - RAZÃO SOCIAL E04 0,64 16,38 2,92 4,85
   - CNPJ E02 0,64 5,84 19,30 4,85 Negrito
   - DATA DA EMISSÃO B09 0,64 4,32 25,14 4,85
   - ENDEREÇO E06 0,64 12,45 2,92 5,49 E07
   - BAIRRO/DISTRITO E09 0,64 5,84 15,37 5,49
   - CEP E13 0,64 3,94 21,21 5,49
   - DATA DA ENTRADA/SAÍDA B10 0,64 4,32 25,14 5,49 Negrito
   - MUNICÍPIO E11 0,64 10,03 2,92 6,13
   - FONE/FAX E16 0,64 5,08 12,95 6,13
   - UF E12 0,64 1,27 18,03 6,13
   - INSCRIÇÃO ESTADUAL E03 0,64 5,84 19,30 6,13
   - FATURA Y02 0,64 26,54 2,92 6,77 Obs FATURA/DUPLICATAS 0,64 0,51 2,41 6,77 Invisível
   - BASE DE CÁLCULO DO ICMS W03 0,64 5,33 2,92 7,41 CÁLCULO DO IMPOSTO 1,28 0,51 2,41 7,41 Invisível
   - VALOR DO ICMS W04 0,64 5,33 8,25 7,41
   - BASE DE CÁLCULO DO ICMS ST W05 0,64 5,33 13,58 7,41
   - VALOR DO ICMS ST W06 0,64 5,33 18,91 7,41
   - VALOR TOTAL DOS PRODUTOS W07 0,64 5,21 24,24 7,41
   - VALOR DO FRETE W08 0,64 4,32 2,92 8,05
   - VALOR DO SEGURO W09 0,64 4,32 7,24 8,05
   - DESCONTO W10 0,64 4,32 11,56 8,05
   - OUTRAS DESPESAS ACESSÓRIAS W15 0,64 4,32 15,88 8,05
   - VALOR DO IPI W12 0,64 4,32 20,20 8,05
   - VALOR TOTAL DA NOTA W16 0,64 4,95 24,52 8,05 Negrito
   - TRANSPORTADOR/VOLUMES TRANSPORTADOS 1,92 0,51 2,41 8,
   - RAZÃO SOCIAL X06 0,64 11,56 2,92 8,69
   - FRETE POR CONTA DE 0,64 2,79 14,48 8,69 Obs
   - CÓDIGO ANTT X21 0,64 2,54 17,27 8,69 X25
   - PLACA DO VEÍCULO X19 0,64 3,81 19,81 8,69 X23
   - UF X20 0,64 1,02 23,62 8,69 X24
   - CNPJ/CPF X04 0,64 4,83 24,64 8,69
   - ENDEREÇO X08 0,64 11,56 2,92 9,33
   - MUNICÍPIO X09 0,64 9,14 14,48 9,33
   - UF X10 0,64 1,02 23,62 9,33
   - INSCRIÇÃO ESTADUAL X07 0,64 4,83 24,64 9,33
   - QUANTIDADE DE VOLUMES X27 0,64 3,56 2,92 9,97
   - ESPÉCIE X28 0,64 3,81 6,48 9,97
   - MARCA X29 0,64 4,19 10,29 9,97
   - NUMERAÇÃO X30 0,64 5,08 14,48 9,97
   - PESO BRUTO X32 0,64 5,08 19,56 9,97
   - PESO LÍQUIDO X31 0,64 4,83 24,64 9,97
   - DADOS DOS PRODUTOS/SERVIÇOS 6,67 0,51 2,41 10,
   - QUADRO DADOS DOS PRODUTOS/SERVIÇOS 6,67 26,54 2,92 10,61 Obs
   - CÓDIGO I02
   - DESCRIÇÃO DOS PRODUTOS/SERVIÇOS I04
   - "COLUNAS ESPECÍFICAS DA EMPRESA" Obs
   - NCM/SH I05
   - CST N11 N
   - CFOP I08
   - UNIDADE I09 I13
   - QUANTIDADE I10 I14
   - VALOR UNITÁRIO I10a I14a
   - DESCONTO I17
   - VALOR TOTAL I11 Obs
   - B.CÁLC.ICMS N15
   - B.CÁLC.ICMS ST N21
   - VALOR ICMS N17
   - VALOR ICMS ST N23
   - VALOR IPI O14
   - ALÍQUOTA ICMS N16
   - ALÍQUOTA IPI O13
   - CÁLCULO DO ISSQN 0,67 0,51 2,41 17,
   - INSCRIÇÃO MUNICIPAL C19 0,67 6,60 2,92 17,28
   - VALOR TOTAL DOS SERVIÇOS W18 0,67 6,60 9,52 17,28
   - BASE DE CÁLCULO DO ISSQN W19 0,67 6,60 16,12 17,28 U02
   - VALOR DO ISSQN W20 0,67 6,73 22,72 17,28 U04
   - DADOS ADICIONAIS 2,94 0,51 2,41 17,
   - INFORMAÇÕES COMPLEMENTARES Z02 2,94 19,05 2,92 17,95 Z03
   - RESERVADO AO FISCO 2,94 7,49 21,97 17, RESERVADO AO FISCO
- Obs 1: Permite-se a inclusão dos dados de duplicatas das TAG do grupo Y


MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

```
Obs 2: Detalhamento específ icos de produtos/serviços (outras TAG do grupo H)
Obs 3: Total Bruto sem desconto
Obs 4: Colunas apresentadas na ordem descrita
Obs 5: TAG: C03, C04, C06, C07, C08, C09, C11, C12, C13, C
Obs 6: TAG: B
Obs 7: TAG: B07, B
Obs 8: TAG: X
Obs 9: Campo utilizado exclusivamente no Modelo de Contingência
```

MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

### 3.9. Campos de Conteúdo Variável

O leiaute de impressão DANFE prevê dois campos de conteúdo variável logo abaixo do local onde é

impressa a chave de acesso, de acordo com a seguinte disposição:

```
DANFE
DOCUMENTO AUXILIAR DA
NOTA FISCAL ELETRÔNICA
0 - ENTRADA
1 - SAÍDA
```
```
Nº 999.999.
SÉRIE 999
FOLHA 01/
```
9999 9999 9999 9999 9999 9999 9999 9999 9999 9999 9999

Campo 1 de conteúdo variável

Campo 2 de conteúdo variável

O conteúdo destes campos é função da forma de emissão da NF-e.

#### 3.9.1. Emissão Normal da NF-e e SVC-XX

A emissão de NF-e normal e a emissão com a utilização da Sefaz Virtual de Contingência do Ambiente

Nacional (SVC-AN) ou da Sefaz Virtual de Contingência do RS (SVC-RS) são formas conclusivas de

emissão da NF-e, pois é dada a autorização de uso para a NF-e, sem necessidade de posterior

transmissão para a SEFAZ.

Nestes casos, após a obtenção da autorização de uso da NF-e o emissor poderá imprimir o DANFE

em papel comum, informando o número do protocolo de autorização de uso e a data e a hora de

autorização no Campo 2, de acordo com a seguinte disposição:

```
DANFE
DOCUMENTO AUXILIAR DA
NOTA FISCAL ELETRÔNICA
```
- ENTRADA
1 - SAÍDA

```
Nº 999.999.
SÉRIE 999
FOLHA 01/
```
9999 9999 9999 9999 9999 9999 9999 9999 9999 9999 9999

```
Consulta de autenticidade no portal nacional da NF-e
http://www.nfe.fazenda.gov.br/portal ou no site da Sefaz Autorizadora
```
11090123456789 12/03/2009 10:00:

O Campo 1 conterá a mensagem informando onde pode ser consultada a autenticidade da NF-e a

partir do valor da chave de acesso.

#### 3.9.2. Emissão da NF-e em Contingência com Impressão do DANFE em Formulário de Segurança...

O uso do formulário de segurança (FS ou FS-DA) para impressão do DANFE é a forma de contingência

mais simples. As NF-e devem ser transmitidas posteriormente para a SEFAZ quando cessados os

problemas técnicos que impediam a transmissão.

```
1
CHAVE DE ACESSO
```
```
1
CHAVE DE ACESSO
```
##### PROTOCOLO DE AUTORIZAÇÃO DE USO


MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

Neste caso, o emissor deverá gerar o Código de Barras Adicional “Dados da NF-e” no Campo 1 e a

representação numérica deste Código de Barras Adicional no Campo 2:

```
DANFE
DOCUMENTO AUXILIAR DA
NOTA FISCAL ELETRÔNICA
0 - ENTRADA
1 - SAÍDA
```
```
Nº 999.999.999
SÉRIE 999
FOLHA 01/01
```
9999 9999 9999 9999 9999 9999 9999 9999 9999 9999 9999

9999 9999 9999 9999 9999 9999 9999 9999 9999

O Código de Barras Adicional dos Dados da NF-e será formado pelo seguinte conteúdo, em um total

de 36 caracteres:

```
cUF tpEmis CNPJ vNF ICMSp ICMSs DD DV
Quantidade de caracteres 02 01 14 14 01 01 02 01
```
- cUF = Código da UF do destinatário ou remetente do Documento Fiscal, informar 99 quando a

operação for de comércio exterior;

- tpEmis = Forma de Emissão da NF-e, informar 2-Contingência FS ou 5-Contingência FS-DA,

conforme o Capítulo 2 do Anexo I do MOC 7.

- CNPJ = CNPJ do destinatário ou do remetente, informar zeros no caso de operação com o exterior

ou o CPF caso o destinatário ou remetente seja pessoa física;

- vNF = Valor Total da NF-e (sem ponto decimal, informar sempre os centavos);
- ICMSp = Destaque de ICMS próprio na NF-e no seguinte formato:

```
o 1 = há destaque de ICMS próprio;
o 2 = não há destaque de ICMS próprio.
```
- ICMSs = Destaque de ICMS por substituição tributária na NF-e, no seguinte formato:

```
o 1 = há destaque de ICMS por substituição tributária;
o 2 = não há destaque de ICMS por substituição tributária.
```
- DD = Dia da emissão da NF-e;
- DV = Dígito Verificador, calculado de forma igual ao DV da Chave de Acesso (item 5.4).

Obs. Todos os campos que formam o código de barras devem ser preenchidos com

alinhamento à direita, sem formatação e com os zeros não significativos necessários para

alcançar o tamanho do campo.

#### 3.9.3. Emissão da NF-e com Prévio Registro do EPEC no Ambiente Nacional

Nesta modalidade de contingência eletrônica o emissor deve gerar o Evento Prévio de Emissão em

Contingência (EPEC), que consiste em um arquivo de resumo das operações que está realizando.

Este arquivo será transmitido ao Ambiente Nacional para autorização do EPEC.

Após o registro do EPEC o emissor poderá imprimir o DANFE em papel comum devendo consignar o

número e data e hora do protocolo de autorização do EPEC no campo 2:

```
DANFE
DOCUMENTO AUXILIAR DA
NOTA FISCAL ELETRÔNICA
0 - ENTRADA
1 - SAÍDA
```
```
Nº 999.999.999
SÉRIE 999
FOLHA 01/01
```
9999 9999 9999 9999 9999 9999 9999 9999 9999 9999 9999

```
Consulta de autenticidade no portal da NF-e
http://www.nfe.fazenda.gov.br/portal
```
11090123456789 12/03/2009 10:00:00

```
1
CHAVE DE ACESSO
```
##### DADOS DA NF-E

```
1
CHAVE DE ACESSO
```
##### PROTOCOLO DE AUTORIZAÇÃO DO EPEC


MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

### 3.10. Outros

#### 3.10.1. Marca d’Água

O formulário poderá conter marca d’água desde que não prejudique a legibilidade dos dados

impressos.

#### 3.10.2. Impressão do Número da Folha

O número de ordem e o número total de folhas deverão ser impressos na parte superior de cada uma

das folhas do DANFE, inclusive na primeira, mesmo que se utilize uma única folha.

### 3.10.3. Limitações da Impressora

Se, no formato retrato, for necessária a utilização de uma margem superior ou inferior maior, devido a

limitações da impressora, a redução necessária poderá ser feita somente na altura do quadro de

“Dados dos Produtos/Serviços” deslocando os campos seguintes para cima pelo valor desta redução.

Essa redução não é permitida no formato paisagem.

#### 3.10.4. Código de Barras

É permitida a impressão de código de barras de informações existentes na NF-e de interesse do

emissor no quadro de informações complementares, no rodapé ou no verso do DANFE.

#### 3.10.5. Campo “Valor de ICMS Desonerado”

O conteúdo do campo vICMSDeson, enquanto não for previsto no leiaute do DANFE, deverá ser

copiado no campo de Informações Complementares de Interesse do Contribuinte (infCpl) para que a

informação conste impressa no DANFE.

Caso seja necessária sua impressão no DANFE, outros campos que não forem previstos no leiaute

também poderão ser copiados no campo de Informações Complementares de Interesse do

Contribuinte (infCpl).

### 3.11. DANFE Simplificado

Nas operações realizadas fora do estabelecimento o DANFE poderá ser impresso em formato

simplificado, não sendo admitida a emissão em contingência utilizando EPEC ou a impressão de

DANFE em formulário de segurança.

#### 3.11.1. Tipo e tamanho do Papel......................................................................................................

Para a impressão do DANFE Simplificado poderá ser utilizado qualquer tipo de papel com largura

mínima de 55 milímetros, com exceção de papel jornal, desde que seja garantido o contraste

necessário para assegurar leitura dos códigos de barras sem problemas.

#### 3.11.2. Chave de acesso

A chave de acesso e seu respectivo código de barras poderão ser impressos em qualquer sentido, no

canto superior direito do papel, observadas as demais disposições dos capítulos 2 e 3 deste Anexo.

#### 3.11.3. Padrões de Caracteres (Tipos de Fontes)

Todos os caracteres deverão estar impressos em tamanho não inferior a seis (6) pontos, sendo os

títulos dos campos impressos em negrito e em caixa alta (maiúsculas).


MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

#### 3.11.4. Campos obrigatórios

No DANFE Simplificado deverão ser impressos, no mínimo, além da expressão “DANFE Simplificado”,

da chave de acesso, seu código de barras e do correspondente Protocolo de Autorização de Uso, o

conteúdo dos seguintes campos:

a) Dados do emitente: Nome/Razão Social, Sigla da UF, CNPJ, Inscrição Estadual;

b) Dados gerais da NF-e: Tipo de operação (entrada ou saída), Série e número da NF-e, Data de

emissão;

c) Dados do destinatário/remetente: Nome/Razão Social, Sigla da UF, CNPJ/CPF;

d) Dados dos itens: Descrição dos Produtos/Serviços, Unidade Comercial, Quantidade, Valor unitário,

Valor total do item;

e) Dados dos totais da NF-e: Valor total da Nota Fiscal.

### 3.12. DANFE Simplificado – Etiqueta (NT 2020.004)

Com o avanço do comércio eletrônico, surgiu a necessidade de simplificar o processo de impressão

do Documento Auxiliar da Nota Fiscal Eletrônica.

A impressão do DANFE Simplificado – Etiqueta, possível de ser utilizado pelos contribuintes nas

operações de venda a varejo para consumidor final em comércio eletrônico, venda por telemarketing

ou processos semelhantes, ocorrerá seguindo os padrões técnicos estabelecidos nesta Nota Técnica,

atendendo ao disposto no §5º-A da cláusula nona do Ajuste SINIEF 07/05.

#### 3.12.1. Tipo e tamanho do Papel......................................................................................................

Para a impressão do DANFE Simplificado poderá ser utilizado qualquer tipo de papel com largura

mínima de 55 milímetros, com exceção de papel jornal, desde que seja garantido o contraste

necessário para assegurar leitura do código de barras nos equipamentos normais do mercado.

#### 3.12.2. Chave de acesso

A chave de acesso e seu respectivo código de barras poderão ser impressos em qualquer sentido, no

canto superior direito do papel, observadas as demais disposições dos capítulos 2 e 3 deste Anexo.

#### 3.12.3. Padrões de Caracteres (Tipos de Fontes)

Todos os caracteres deverão estar impressos em tamanho não inferior a seis (6) pontos, sendo os

títulos dos campos impressos em negrito e em caixa alta (maiúsculas).

#### 3.12.4. Campos obrigatórios

No DANFE Simplificado – Etiqueta deverão estar visíveis e ser impresso no mínimo, além da chave

de acesso, seu código de barras e do correspondente Protocolo de Autorização de Uso, o conteúdo

dos seguintes campos:

a) A descrição “DANFE Simplificado – Etiqueta”;

b) Dados do emitente: Nome/Razão Social, Sigla da UF, CNPJ, Inscrição Estadual;

c) Dados gerais da NF-e: Tipo de operação, se entrada ou saída, Série e Número da NF-e, Data de

emissão;

d) Dados do destinatário/remetente: Nome/Razão Social, Sigla da UF, CNPJ/CPF, Inscrição

Estadual, quando existir;

e) Dados dos totais da NF-e: Valor total da Nota Fiscal.

f) Contingência EPEC: Informar o protocolo de autorização do Evento EPEC.


MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

## Anexo III.01 – Conjunto de Caracteres Código de Barras CODE-128C

```
Combinação de barras: B = barra preta e S = espaço (barra branca)
Valor Valor Valor
CODE C B S B S B S CODE C B S B S B S CODE C B S B S B S
00 2 1 2 2 2 2 50 2 3 1 1 3 1 100 1 1 4 1 3 1
01 2 2 2 1 2 2 51 2 1 3 1 1 3 101 3 1 1 1 4 1
02 2 2 2 2 2 1 52 2 1 3 3 1 1 102 4 1 1 1 3 1
03 1 2 1 2 2 3 53 2 1 3 1 3 1 103 2 1 1 4 1 2
04 1 2 1 3 2 2 54 3 1 1 1 2 3 104 2 1 1 2 1 4
05 1 3 1 2 2 2 55 3 1 1 3 2 1
06 1 2 2 2 1 3 56 3 3 1 1 2 1
07 1 2 2 3 1 2 57 3 1 2 1 1 3
08 1 3 2 2 1 2 58 3 1 2 3 1 1
09 2 2 1 2 1 3 59 3 3 2 1 1 1
10 2 2 1 3 1 2 60 3 1 4 1 1 1
11 2 3 1 2 1 2 61 2 2 1 4 1 1
12 1 1 2 2 3 2 62 4 3 1 1 1 1
13 1 2 2 1 3 2 63 1 1 1 2 2 4
14 1 2 2 2 3 1 64 1 1 1 4 2 2
15 1 1 3 2 2 2 65 1 2 1 1 2 4
16 1 2 3 1 2 2 66 1 2 1 4 2 1
17 1 2 3 2 2 1 67 1 4 1 1 2 2
18 2 2 3 2 1 1 68 1 4 1 2 2 1
19 2 2 1 1 3 2 69 1 1 2 2 1 4
20 2 2 1 2 3 1 70 1 1 2 4 1 2
21 2 1 3 2 1 2 61 1 2 2 1 1 4
22 2 2 3 1 1 2 72 1 2 2 4 1 1
23 3 1 2 1 3 1 73 1 4 2 1 1 2
24 3 1 1 2 2 2 74 1 4 2 2 1 1
25 3 2 1 1 2 2 75 2 4 1 2 1 1
26 3 2 1 2 2 1 76 2 2 1 1 1 4
27 3 1 2 2 1 2 77 4 1 3 1 1 1
28 3 2 2 1 1 2 78 2 4 1 1 1 2
29 3 2 2 2 1 1 79 1 3 4 1 1 1
30 2 1 2 1 2 3 80 1 1 1 2 4 2
31 2 1 2 3 2 1 81 1 2 1 1 4 2
32 2 3 2 1 2 1 82 1 2 1 2 4 1
33 1 1 1 3 2 3 83 1 1 4 2 1 2
34 1 3 1 1 2 3 84 1 2 4 1 1 2
35 1 3 1 3 2 1 85 1 2 4 2 1 1
36 1 1 2 3 1 3 86 4 1 1 2 1 2
37 1 3 2 1 1 3 87 4 2 1 1 1 2
38 1 3 2 3 1 1 88 4 2 1 2 1 1
39 2 1 1 3 1 3 89 2 1 2 1 4 1
40 2 3 1 1 1 3 90 2 1 4 1 2 1
41 2 3 1 3 1 1 91 4 1 2 1 2 1
42 1 1 2 1 3 3 92 1 1 1 1 4 3
43 1 1 2 3 3 1 93 1 1 1 3 4 1
44 1 3 2 1 3 1 94 1 3 1 1 4 1
45 1 1 3 1 2 3 95 1 1 4 1 1 3
46 1 1 3 3 2 1 96 1 1 4 3 1 1
47 1 3 3 1 2 1 97 4 1 1 1 1 3
48 3 1 3 1 2 1 98 4 1 1 3 1 1
49 2 1 1 3 3 1 99 1 1 3 1 4 1
```
```
Valor
B S B S B S B S B S B S B
2 1 1 2 3 2 2 3 3 1 1 1 2
```
```
Combinação de Barras
```
```
Conjunto de caracteres representativos do Código de Barras CODE-128C
```
```
Caractere de Início (START)
105
```
```
Caractere de Fim (STOP)
```
```
Combinação de Barras Combinação de Barras
```

MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

## Anexo III.02 – DANFE Tamanho A-4 em Modo Retrato, Folhas Soltas


MOC 7.0 **–** Anexo II, Manual de Especificações Técnicas do DANFE e Código de Barras

## Anexo III.03 – DANFE Tamanho A-4 em Modo Retrato, Formulário Contínuo


Projeto

Nota Fiscal Eletrônica
MOC **–** Manual do DANFE

## Anexo III.04 – DANFE Tamanho A-4 em Modo Paisagem, Folhas Soltas............................................


Projeto

Nota Fiscal Eletrônica
MOC **–** Manual do DANFE

## Anexo III.05 - DANFE Tamanho A-4 em Modo Paisagem, Formulário Contínuo



