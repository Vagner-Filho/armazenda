# DANFE — Regras e Especificações para Geração

Resumo das regras obrigatórias que o sistema deve seguir para gerar o DANFE
(Documento Auxiliar da Nota Fiscal Eletrônica) conforme o **MOC 7.0 – Anexo II**
(`docs/moc_layout_danfe.md`). Este documento é a referência de implementação e
auditoria de conformidade do gerador de DANFE do Armazenda.

Notas de escopo:
- NF-e **modelo 55** (não NFC-e), emissores pessoa física (CPF) e jurídica (CNPJ).
- Produtos restritos ao agronegócio.
- Contingências consideradas: normal, SVC-AN/SVC-RS, FS-DA e EPEC (dados no
  modelo mas apenas SVC é ativamente ligada — ver `AGENTS.md`).

Referências rápidas:
- Layout XML (Anexo I): `docs/moc_layout.md`
- Contingência (Anexo IV): `docs/moc_contingencia.md`
- Manual original: `docs/moc_layout_danfe.md`

---

## 1. Visão geral do DANFE

O DANFE é um documento auxiliar **impresso** que acompanha mercadorias em trânsito
e serve para:

- Acompanhar o trânsito de mercadorias
- Colher firma do destinatário/tomador (comprovação de entrega)
- Prover representações impressas previstas em legislação
- Auxiliar a escrituração por destinatários não credenciados

Regras gerais (§3 da Anexo II):

- Uma única via, salvo disposição expressa.
- Papel qualquer, exceto jornal; contraste suficiente para leitura do código de barras.
- **Não pode imprimir informação que não conste do XML da NF-e** (§3.1).
- Homologação: imprimir **"SEM VALOR FISCAL"** em Informações Complementares ou
  como marca d'água (§3).
- Contingência: destacar essa condição (Anexo IV).
- "Valor Aproximado dos Tributos" é opcional — podevir em coluna própria no
  Quadro de Cálculo do Imposto e/ou no Quadro de Dados dos Produtos/Serviços,
  além de em `infAdProd`/`infCpl` (§3 intro).

---

## 2. Código de barras — CODE-128C (§2)

Padrão: **CODE-128C**.

Quantidade de códigos de barras a imprimir:

| Forma de emissão | Códigos de barras | Conteúdo |
|---|---|---|
| Normal / SVC-AN / SVC-RS | 1 | Chave de Acesso (44 dígitos) |
| FS / FS-DA | 2 | Chave de Acesso + "Dados da NF-e" (36 dígitos) |
| EPEC | 2 | Chave de Acesso + campo do protocolo EPEC |

Dimensões mínimas (Chave de Acesso, 44 posições):

| Parâmetro | Mínimo |
|---|---|
| Largura total — Laser/Jato de tinta | 6,0 cm |
| Largura total — Matricial/Linha | 11,5 cm |
| Altura da barra | 0,8 cm |
| Largura de cada módulo | 0,02 cm |

Estrutura simbólica:
```
Margem clara | Start C | Dados | DV | Stop | Margem clara
```

- DV do CODE-128C: **módulo 103**, soma ponderada (1×Start=105 + 2×dígito1 +
  3×dígito2 + ...), resto da divisão por 103.
- Tabela de caracteres CODE-128C: Anexo III.01 de `moc_layout_danfe.md`.

### 2.1 Conteúdo do Código de Barras Adicional "Dados da NF-e" (FS/FS-DA)

36 caracteres, alinhamento à direita com zeros não significativos:

| Campo | Tam | Descrição |
|---|---|---|
| cUF | 2 | UF do remetente/destinatário (99 p/ exterior) |
| tpEmis | 1 | 2 (FS) ou 5 (FS-DA) |
| CNPJ | 14 | CNPJ do destinatário/remetente (zeros p/ exterior; CPF se PF) |
| vNF | 14 | Valor total da NF-e sem ponto, sempre com centavos |
| ICMSp | 1 | 1 = há destaque ICMS próprio; 2 = não há |
| ICMSs | 1 | 1 = há destaque ICMS ST; 2 = não há |
| DD | 2 | Dia da emissão |
| DV | 1 | Dígito verificador (mesmo cálculo da Chave de Acesso, módulo 11 base 2,9) |

Exibição numérica (Campo 2) no DANFE: **9 blocos de 4 dígitos**.

---

## 3. Chave de Acesso (§3.1.1)

Impressa em **11 blocos de 4 dígitos**:

```
9999 9999 9999 9999 9999 9999 9999 9999 9999 9999 9999
```

Conteúdo do campo: **negrito** (§3.7.5).

---

## 4. Quadros obrigatórios e seus campos

Cada quadro do DANFE reflete o conteúdo de TAGs específicas do XML. Lista completa:

### 4.1 Quadro Identificação do Emitente (Grupo C)
- Nome/Razão Social (C03)
- Endereço completo: logradouro (C06), número (C07), complemento (C08),
  bairro (C09), município (C11), UF (C12), CEP (C13)
- Telefone (C16)
- Logo opcional, desde que não prejudique as informações obrigatórias.

### 4.2 Cabeçalho "DANFE"
- "DANFE" em destaque
- "DOCUMENTO AUXILIAR DA NOTA FISCAL ELETRÔNICA"
- Tipo de operação: "0 - ENTRADA" / "1 - SAÍDA" (tag tpNF, B11)
- Número (B08) e série (B07) da NF-e
- Folha "nn/total" (em todas as folhas, inclusive a primeira)

### 4.3 Código de Barras da Chave
- Conforme §2.

### 4.4 Campo de Conteúdo Variável (Campos 1 e 2)
- Conforme §7 deste documento (depende de `tpEmis`).

### 4.5 Natureza da Operação / chave de acesso
- Natureza da Operação (B04)
- Chave de Acesso (A03) — já tratada acima

### 4.6 Inscrições e CNPJ do Emitente
- IE do Emitente (C17)
- IE ST do Emitente (C18)
- CNPJ do Emitente (C02)

### 4.7 Destinatário/Remetente (Grupo E)
| Campo | TAG | Tam. |
|---|---|---|
| Razão Social | E04 | 60 |
| CNPJ/CPF | E02/E03 | 14/11 |
| Data da Emissão | B09 | 10 |
| Endereço | E06 + E07 | 120 |
| Bairro/Distrito | E09 | 60 |
| CEP | E13 | 8 |
| Data Entrada/Saída | B10 | 10 |
| Município | E11 | 60 |
| Fone/Fax | E16 | 10 |
| UF | E12 | 2 |
| Inscrição Estadual | E17/E03 | 14 |
| Hora Entrada/Saída | — | — |

### 4.8 Fatura/Duplicatas (Grupo Y)
- Opcional/suprimível (§6.2). Pode incluir outras informações do Grupo Y desde
  que estejam no XML.

### 4.9 Cálculo do Imposto (Grupo W)
| Campo | TAG |
|---|---|
| Base de Cálculo ICMS | W03 |
| Valor do ICMS | W04 |
| Base de Cálculo ICMS ST | W05 |
| Valor do ICMS ST | W06 |
| Valor Total dos Produtos | W07 |
| Valor do Frete | W08 |
| Valor do Seguro | W09 |
| Desconto | W10 |
| Outras Despesas Acessórias | W15 |
| Valor do IPI | W12 |
| Valor Total da Nota | W16 (negrito) |

### 4.10 Transportador/Volumes Transportados (Grupo X)
TAGs: `modFrete` (X02), RNTRC/ANTT (X25), placa (X19/X23), UF do veículo (X10),
CNPJ/CPF transportador (X04), endereço (X08), município (X09), UF (X10),
IE (X07), quantidade de volumes (X27), espécie (X28), marca (X29), numeração
(X30), peso bruto (X32), peso líquido (X31).

`modFrete` valores (NT 2018.005):
```
0 = CIF (Remetente)
1 = FOB (Destinatário)
2 = Terceiros
3 = Transporte Próprio Remetente
4 = Transporte Próprio Destinatário
9 = Sem Ocorrência de Transporte
```

### 4.11 Quadro Dados dos Produtos/Serviços
Ver §5 (regras detalhadas de colunas).

### 4.12 Cálculo do ISSQN (Grupo U)
Suprimível se não aplicável (§6.3). TAGs: C19 (IM), W18, W19 (U02), W20 (U04).

### 4.13 Dados Adicionais
- **Informações Complementares** (Z02/Z03): `infAdFisco` + `infCpl`. Pode continuar
  no verso ou na folha seguinte, no mesmo quadro ou no de Dados dos Produtos.
- **Reservado ao Fisco**: o contribuinte **não preenche** (uso exclusivo do fisco).

### 4.14 Local de Retirada (Grupo F) e Local de Entrega (Grupo G)
Exibição facultada em área específica (NT 2018.005).

---

## 5. Quadro "Dados dos Produtos/Serviços" (§3.1.7, §3.2)

### 5.1 Colunas não suprimíveis (obrigatórias)
- Código do Produto/Serviço (I02)
- Descrição do Produto/Serviço (I04)
- NCM (I05)
- CST (N11)
- CFOP (I08)
- Unidade (I09/I13)
- Quantidade (I10/I14)
- Valor Unitário (I10a/I14a)
- Valor Total (I11)
- Base de Cálculo do ICMS próprio (N15)
- Valor do ICMS próprio (N17)
- Alíquota do ICMS (N16)

Outras colunas podem ser suprimidas; outras podem ser **acrescentadas à direita**
da coluna "Descrição", respeitando a ordem das remanescentes.

### 5.2 Colunas que podem ser combinadas numa mesma coluna (§3.2)
Para usar duas linhas por item:
- Código do Produto + NCM/SH
- CST + CFOP
- CSOSN + CFOP
- Quantidade + Unidade
- Valor Unitário + Desconto
- Valor Total + Base de Cálculo ICMS
- BC ICMS ST + Valor ICMS ST
- Valor ICMS Próprio + Valor IPI
- Alíquota ICMS + Alíquota IPI

A coluna "Descrição dos Produtos/Serviços" **nunca** pode ser combinada com outra.

### 5.3 Divisores entre itens
Cada item deve ser destacado do seguinte por:
- Linha tracejada/contínua/tracejada, ou
- Espaçamento duplo, ou
- Sombreamento, ou
- Recurso semelhante que resulte em destaque divisório claro.

### 5.4 Informações adicionais de produto
- `infAdProd` (V01) imprimida imediatamente abaixo do item a que se refere.
- Valores de FCP por item (`vBCFCP`, `pFCP`, `vFCP`, `vBCFCPST`, `pFCPST`,
  `vFCPST`) vão em `infAdProd`.
- vUnCom ≠ vUnTrib → ambos devem aparecer identificados (linha extra ou
  inf. adicional).

---

## 6. Modificações e supressões permitidas (§3.3)

| Bloco | Suprimível? | Observação |
|---|---|---|
| Canhoto | Sim, somente no formato retrato | Recuperar espaço para o quadro de Produtos (deslocar campos seguintes para cima) |
| Fatura/Duplicatas | Sim | Reduz/e/ou elimina; altura recuperada vai para Dados dos Produtos |
| Cálculo do ISSQN | Sim | Recuperação dividida entre Produtos e Inf. Complementares + Reservado ao Fisco |

Para DANFE que não usa formulário de segurança, o canhoto pode ser deslocado para
a extremidade inferior sem alterar demais dimensões (somente retrato).

---

## 7. Campos de Conteúdo Variável (§3.9)

Dois campos impressos logo abaixo da Chave de Acesso. Conteúdo depende de `tpEmis`:

### 7.1 Emissão Normal e SVC-XX (tpEmis 1, 6, 7)
- Campo 1: mensagem de consulta de autenticidade
  (`Consulta de autenticidade no portal nacional da NF-e
  http://www.nfe.fazenda.gov.br/portal ou no site da Sefaz Autorizadora`).
- Campo 2: número do protocolo de autorização de uso + data/hora autorização.

### 7.2 Emissão em Contingência FS ou FS-DA (tpEmis 2, 5)
- Campo 1: **Código de barras adicional "Dados da NF-e"** (36 caracteres — ver §2.1).
- Campo 2: **representação numérica** do código de barras adicional, em 9 blocos
  de 4 dígitos: `9999 9999 9999 9999 99 99 9999 9999 9999`.

### 7.3 Emissão EPEC (tpEmis 4)
- Campo 1: mensagem de consulta de autenticidade.
- Campo 2: número do protocolo de autorização do EPEC + data/hora.

---

## 8. Verso e folhas adicionais (§3.4, §3.5)

### 8.1 Verso
- Até **50%** pode continuar os quadros "Dados dos Produtos/Serviços" e/ou
  "Informações Complementares"; o restante deve ficar sem impressão.
- Imprimir "**CONTINUA NO VERSO**" no anverso, ao final do(s) quadro(s) que
  continua(m).

### 8.2 Folhas adicionais
- O DANFE pode ter várias folhas.
- Toda folha adicional deve repetir no topo, no mínimo:
  - Identificação do Emitente
  - "DANFE" e "DOCUMENTO AUXILIAR DA NOTA FISCAL ELETRÔNICA"
  - Número e série da NF-e, tipo de operação, "Folha nn/total"
  - Código(s) de barras
  - Natureza da Operação e Chave de Acesso
  - IE, IE-ST e CNPJ do Emitente
- A área restante é usada exclusivamente para:
  - Continuação dos itens (mesmas colunas/larguras da 1ª folha)
  - Continuação das informações complementares

---

## 9. Formulário (§3.6)

- Tamanho do papel: **A4 (mínimo)** até **Ofício II (230×330 mm) (máximo)**.
- Espaço excedente: horizontalmente aumenta largura dos campos; verticalmente
  aumenta altura **apenas** dos quadros:
  - Dados dos Produtos/Serviços, ou
  - Informações Complementares + Reservado ao Fisco, ou
  - Combinação desses.
- Margens laterais (esquerda/direita/superior/inferior): **0,2 cm a 0,8 cm**.
- Modelos permitidos (Anexos III.02–III.05):
  - A4 retrato — folhas soltas ou formulário contínuo
  - A4 paisagem — folhas soltas ou formulário contínuo
- Disposição dos campos deve obedecer ao modelo correspondente.

### 9.1 Limitações de impressora
- No retrato, se for necessária margem superior/inferior maior, a redução pode
  ser feita **somente** na altura do quadro de Produtos, deslocando os campos
  seguintes para cima.
- Não permitido no formato paisagem (§3.10.3).

---

## 10. Padrões de caracteres (fontes) — §3.7

Fontes aceitas: **Times New Roman** ou **Courier New**. Tamanhos mínimos:

| Elemento | Mínimo | Estilo |
|---|---|---|
| Descritivo dos blocos de campos | 5 pt | Negrito, caixa alta |
| Descritivo dos campos do quadro de Produtos | 5 pt | Caixa alta |
| Descritivo dos demais campos | 6 pt | Caixa alta |
| "DANFE" | 12 pt (ou 10 CPP) | Negrito, caixa alta |
| Série/número/folha/tipo operação | 10 pt (ou 10 CPP) | Negrito, caixa alta |
| "DOCUMENTO AUXILIAR DA NF-E" + ENTRADA/SAÍDA | 8 pt (ou 17 CPP) | caixa alta |
| Chave de Acesso (conteúdo) | — | Negrito |
| Razão Social e/ou nome fantasia do emitente | 12 pt (ou 17 CPP) | Negrito |
| Demais dados do emitente (endereço, município, CEP, fone) | 8 pt (ou 17 CPP) | Negrito |
| Conteúdo do quadro Dados dos Produtos | 6 pt (ou 17 CPP) | — |
| Conteúdo de Informações Complementares | 6 pt (ou 17 CPP) | — |
| Conteúdo dos demais campos | 10 pt (ou 17 CPP) | — |

Para impressoras de impacto (matricial/linha): entre **10 e 17 CPP**.

---

## 11. Tamanhos e posições dos campos (§3.8)

Definidos com precisão em cm, eixo zero no canto superior esquerdo, para:

- **§3.8.1** — A4 Retrato (referência em `docs/moc_layout_danfe.md` linhas
  834–1089). Modelo visual completo no Anexo III.02.
- **§3.8.2** — A4 Paisagem (linhas 1090–1307). Modelo visual no Anexo III.04.

Trechos relevantes do formato Retrato (alta prioridade para implementação):

| Bloco/Campo | Altura | Largura | Esq. | Sup. | Tam. TAG | Obs |
|---|---|---|---|---|---|---|
| Canhoto — Recebemos de... | 0,85 | 16,10 | 0,25 | 0,00 | — | |
| Canhoto — NF-e/Nº/Série | 1,70 | 4,50 | 16,35 | 0,00 | — | |
| Quadro Identif. Emitente | 3,00–3,50 | 10,00 | 0,00 | 0,00–2,20 | Obs 5 | TAGs C03,C04,C06..C16 |
| Quadro "DANFE" | 3,30 | 2,20 | 5,00 | 2,20 | — | |
| Quadro Código de Barras da Chave | 1,00 | 12,80 | 8,12 | 2,20 | — | Mat. ou Laser |
| Código de Barras da Chave | 1,00 | 11,50 | 8,62 | 2,20 | — | |
| Chave de Acesso | 0,85 | 12,70 | 8,12 | 4,02 | 44 | |
| Código de Barras dos Dados (conting.) | 1,00 | 12,70 | 8,12 | 4,98 | — | Mat. ou Laser;Obs 9 |
| Natureza da Operação | 0,85 | 7,87 | 0,25 | 6,46 | 60 | TAG B04 |
| IE Emitente | 0,85 | 6,86 | 0,25 | 7,31 | 14 | TAG C17 |
| IE ST Emitente | 0,85 | 6,86 | 7,11 | 7,31 | 14 | TAG C18 |
| CNPJ Emitente | 0,85 | 6,86 | 13,97 | 7,31 | 14 | TAG C02 |
| Razão Social (dest) | 0,85 | 12,32 | 0,25 | 8,58 | 60 | TAG E04 |
| CNPJ (dest) | 0,85 | 5,33 | 12,57 | 8,58 | 14 | TAG E02 (negrito) |
| Data Emissão | 0,85 | 2,92 | 17,90 | 8,58 | 10 | TAG B09 |
| Endereço | 0,85 | 10,16 | 0,25 | 9,43 | 120 | TAG E06/E07 |
| Bairro/Distrito | 0,85 | 4,83 | 10,41 | 9,43 | 60 | TAG E09 |
| CEP | 0,85 | 2,67 | 15,24 | 9,43 | 8 | TAG E13 |
| Data Entrada/Saída | 0,85 | 2,92 | 17,91 | 9,43 | 10 | TAG B10 (negrito) |
| Município | 0,85 | 7,11 | 0,25 | 10,28 | 60 | TAG E11 |
| Fone/Fax | 0,85 | 4,06 | 7,36 | 10,28 | 10 | TAG E16 |
| UF (dest) | 0,85 | 1,14 | 11,42 | 10,28 | 2 | TAG E12 |
| IE (dest) | 0,85 | 5,33 | 12,56 | 10,28 | 14 | TAG E17 |
| Hora Entrada/Saída | 0,85 | 2,92 | 17,89 | 10,28 | — | Negrito |
| Fatura | 0,85 | 20,57 | 0,25 | 11,51 | — | Obs 1 |
| Base Cálculo ICMS | 0,85 | 4,06 | 0,25 | 12,78 | 15 | TAG W03 |
| Valor ICMS | 0,85 | 4,06 | 4,31 | 12,78 | 15 | TAG W04 |
| BC ICMS ST | 0,85 | 4,06 | 8,37 | 12,78 | 15 | TAG W05 |
| Valor ICMS ST | 0,85 | 4,06 | 12,43 | 12,78 | 15 | TAG W06 |
| Valor Total dos Produtos | 0,85 | 4,32 | 16,49 | 12,78 | 15 | TAG W07 |
| Valor Frete | 0,85 | 3,30 | 0,25 | 13,63 | 15 | TAG W08 |
| Valor Seguro | 0,85 | 3,30 | 3,55 | 13,63 | 15 | TAG W09 |
| Desconto | 0,85 | 3,30 | 6,85 | 13,63 | 15 | TAG W10 |
| Outras Despesas | 0,85 | 3,30 | 10,15 | 13,63 | 15 | TAG W15 |
| Valor IPI | 0,85 | 3,30 | 13,45 | 13,63 | 15 | TAG W12 |
| Valor Total da Nota | 0,85 | 4,06 | 16,75 | 13,63 | 15 | TAG W16 (negrito) |
| Razão Social (transp) | 0,85 | 9,02 | 0,25 | 14,90 | 60 | TAG X06 |
| Frete por Conta | 0,85 | 2,79 | 9,27 | 14,90 | — | Obs 8 (TAG X02) |
| Código ANTT | 0,85 | 1,78 | 12,06 | 14,90 | 20 | TAG X25 |
| Placa do Veículo | 0,85 | 2,29 | 13,84 | 14,90 | 8 | TAG X19 |
| UF (transp/veículo) | 0,85 | 0,76 | 16,13 | 14,90 | 2 | TAG X10 |
| CNPJ/CPF (transp) | 0,85 | 3,94 | 16,89 | 14,90 | 14 | TAG X04 |
| Endereço (transp) | 0,85 | 9,02 | 0,25 | 15,75 | 60 | TAG X08 |
| Município (transp) | 0,85 | 6,86 | 9,27 | 15,75 | 60 | TAG X09 |
| UF (transp) | 0,85 | 0,76 | 16,13 | 15,75 | 2 | TAG X10 |
| IE (transp) | 0,85 | 3,94 | 16,89 | 15,75 | 14 | TAG X07 |
| Quantidade Volumes | 0,85 | 2,92 | 0,25 | 16,60 | 15 | TAG X27 |
| Espécie | 0,85 | 3,05 | 3,17 | 16,60 | 60 | TAG X28 |
| Marca | 0,85 | 3,05 | 6,22 | 16,60 | 60 | TAG X29 |
| Numeração | 0,85 | 4,83 | 9,27 | 16,60 | 60 | TAG X30 |
| Peso Bruto | 0,85 | 3,43 | 14,10 | 16,60 | 15 | TAG X32 |
| Peso Líquido | 0,85 | 3,30 | 17,53 | 16,60 | 15 | TAG X31 |
| Quadro Dados dos Produtos | 6,77 | 20,57 | 0,25 | 17,87 | — | Obs 4 |
| Inscrição Municipal | 0,85 | 5,08 | 0,25 | 25,06 | 15 | TAG C19 |
| Valor Total dos Serviços | 0,85 | 5,08 | 5,33 | 25,06 | 15 | TAG W18 |
| BC do ISSQN | 0,85 | 5,08 | 10,41 | 25,06 | 15 | TAG W19 (U02) |
| Valor do ISSQN | 0,85 | 5,33 | 15,49 | 25,06 | 15 | TAG W20 (U04) |
| Informações Complementares | 3,07 | 12,95 | 0,25 | 26,33 | 5256 | TAG Z02/Z03 |
| Reservado ao Fisco | 3,07 | 7,62 | 13,17 | 26,33 | — | |

Observações do Anexo II:
- Obs 1: permite incluir dados de duplicatas (Grupo Y).
- Obs 2: detalhamentos específicos do Grupo H.
- Obs 3: Total Bruto (TAG) ou Líquido (Mod. 1/1-A).
- Obs 4: colunas na ordem descrita.
- Obs 5: TAGs C03, C04, C06, C07, C08, C09, C11, C12, C13, C16.
- Obs 6: TAG B11 (tpNF).
- Obs 7: TAGs B07, B08.
- Obs 8: TAG X02 (modFrete).
- Obs 9: campo exclusivo do modelo de contingência.

Para o formato Paisagem, ver §3.8.2 em `docs/moc_layout_danfe.md`.

---

## 12. Regras diversas (§3.10)

- **Marca d'água** permitida se não prejudicar legibilidade (§3.10.1).
- **Número da folha**: "Folha n/total" impresso no topo de **todas** as folhas,
  inclusive a primeira (§3.10.2).
- **Outros códigos de barras**: pode imprimir códigos de barras adicionais no
  quadro de Inf. Complementares, rodapé ou verso (§3.10.4).
- **vICMSDeson**: enquanto não tiver campo próprio no leiaute do DANFE, deve ser
  copiado para `infCpl` para constar impresso (§3.10.5). Outros campos sem leiaute
  dedicado também podem ser copiados para `infCpl`.

---

## 13. DANFE Simplificado (§3.11)

Uso: operações realizadas **fora do estabelecimento**.

- **Não** admitido em contingência EPEC nem com formulário de segurança.
- Papel: largura mínima **55 mm**, contraste suficiente para leitura do código
  de barras.
- Chave de acesso + código de barras: canto superior direito, qualquer sentido.
- Fontes: mínimas 6 pt; títulos em negrito e caixa alta.

Campos obrigatórios (além da string "DANFE Simplificado", chave de acesso,
código de barras e protocolo de autorização de uso):

- Emitente: Nome/Razão Social, UF, CNPJ, IE
- Gerais da NF-e: tipo de operação (E/S), série, número, data de emissão
- Destinatário: Nome/Razão Social, UF, CNPJ/CPF
- Itens: descrição, unidade comercial, quantidade, valor unitário, valor total
  do item
- Total da NF-e: valor total da Nota Fiscal

---

## 14. DANFE Simplificado — Etiqueta (§3.12, NT 2020.004)

Uso: comércio eletrônico, telemarketing ou processos semelhantes com consumidor
final.

- Papel: largura mínima **55 mm**.
- Chave de acesso + código de barras: canto superior direito, qualquer sentido.
- Fontes: mínimas 6 pt; títulos em negrito e caixa alta.

Campos obrigatórios (além do rótulo "DANFE Simplificado – Etiqueta", chave
de acesso, código de barras e protocolo de autorização de uso):

- Emitente: Nome/Razão Social, UF, CNPJ, IE
- Gerais da NF-e: tipo de operação, série, número, data de emissão
- Destinatário: Nome/Razão Social, UF, CNPJ/CPF, IE quando existir
- Total da NF-e: valor total da Nota Fiscal
- EPEC quando for o caso: protocolo de autorização do EPEC

---

## 15. Checklist de conformidade para o Armazenda

1. Definir `tpImp` (B21) no XML:
   - `1` para retrato (modelo padrão do Armazenda).
   - `3` para DANFE Simplificado (se aplicável).
2. Renderizar o quadro de Identificação do Emitente com todas as TAGs do Grupo C.
3. Renderizar cabeçalho "DANFE" / "DOCUMENTO AUXILIAR…" com tamanhos mínimos
   (12 pt "DANFE" em negrito, 10 pt série/número/folha em negrito, 8 pt
   "DOCUMENTO AUXILIAR…").
4. Marcar tipo de operação "ENTRADA"/"SAÍDA" conforme `tpNF` (B11).
5. Imprimir "Folha nn/total" no topo de toda folha, inclusive a primeira.
6. Imprimir Chave de Acesso formatada em 11 blocos de 4 dígitos, em negrito.
7. Imprimir código de barras CODE-128C da Chave de Acesso com as dimensões
   mínimas de §2.
8. Renderizar Campos 1 e 2 de Conteúdo Variável conforme `tpEmis` (ver §7),
   incluindo:
   - URL de consulta para normal/SVC/EPEC.
   - Código de barras adicional "Dados da NF-e" (36 chars) para FS/FS-DA.
   - Protocolo + data/hora de autorização (ou EPEC).
9. Renderizar quadro do Destinatário/Remetente (Grupo E) com todas as TAGs
   listadas em §4.7.
10. Renderizar quadro Cálculo do Imposto (Grupo W) com W03–W16, com W16 em negrito.
11. Renderizar quadro Transportador/Volumes (Grupo X) com `modFrete` conforme
    tabela em §4.10.
12. No quadro Dados dos Produtos/Serviços, manter todas as colunas insubstituíveis
    de §5.1 e respeitar combinações permitidas (§5.2).
13. Aplicar divisor entre itens (linha tracejada, espaçamento duplo ou sombreamento).
14. Imprimir `infAdProd` abaixo do item correspondente; valores FCP por item
    vão em `infAdProd`; totais FCP em `infAdFisco`.
15. Renderizar quadro de Informações Complementares com `infAdFisco` e `infCpl`;
    copiar `vICMSDeson` para `infCpl` (§3.10.5).
16. Deixar o quadro Reservado ao Fisco não preenchido pelo contribuinte.
17. Para homologação, incluir "SEM VALOR FISCAL".
18. Para contingência, destacar a condição no DANFE conforme Anexo IV.
19. Respeitar supressões/modificações permitidas (Canhoto/Fatura/ISSQN) por formato.
20. Verso: até 50%, com "CONTINUA NO VERSO" no anverso quando usado.
21. Folhas adicionais: repetir identificação mínima no topo e usar área
    restante para itens e/ou inf. complementares.
22. Fonte Times New Roman ou Courier New, com tamanhos mínimos por elemento
    de §10.
23. Tamanhos/posições mínimos do Retrato conforme §11 (ou Paisagem conforme
    §3.8.2).
24. Margens laterais 0,2 cm a 0,8 cm.
25. Papel A4 (mínimo) a Ofício II (máximo).

---

## Referências internas

- Layout NF-e (Anexo I): `docs/moc_layout.md`
- Manual original do DANFE (Anexo II): `docs/moc_layout_danfe.md`
- Manual de Contingência (Anexo III/IV): `docs/moc_contingencia.md`
- Arquitetura NF-e do Armazenda: `AGENTS.md` (seção "NF-e Contingency Architecture")