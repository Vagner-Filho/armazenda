# PL_010d_v1.03 vs PL_010d_v1.01 (PL v1.01) — schema diff

Generated: 2026-08-25T18:43:03

**Old PL**: `docs/Evento_Canc_PL_v1.01/` (Pacote de Liberação v1.01, evento cancelamento — 8 files, last updated 2018-12-19)
**New PL**: `docs/cancel_event_research/tier_1_and_tier_2/PL_010d_v1.03/PL_010d_v1.03/Evento/` (Pacote de Liberação 010d v1.03, NT 2026.004 v.1.01)

---

## 1. File-level diff

Old PL file count: 8
New PL file count: 7
Common files: 2

### Files ADDED in PL_010d_v1.03:
- `envEvento_v1.00.xsd`
- `leiauteEvento_v1.00.xsd`
- `procEventoNFe_v1.00.xsd`
- `retEnvEvento_v1.00.xsd`
- `tiposBasico_v4.00.xsd`

### Files REMOVED in PL_010d_v1.03 (cancellation-specific schemas):
- `e110111_v1.00.xsd`
- `envEventoCancNFe_v1.00.xsd`
- `eventoCancNFe_v1.00.xsd`
- `leiauteEventoCancNFe_v1.00.xsd`
- `procEventoCancNFe_v1.00.xsd`
- `retEnvEventoCancNFe_v1.00.xsd`

### Common files (in both):
- `tiposBasico_v1.03.xsd`
- `xmldsig-core-schema_v1.01.xsd`

---

## 2. Unified diff for common files

### `tiposBasico_v1.03.xsd` — diff

```diff
--- old/tiposBasico_v1.03.xsd
+++ new/tiposBasico_v1.03.xsd
@@ -1,12 +1,3 @@
 <?xml version="1.0" encoding="UTF-8"?>

-<!-- PL_006u - 21/07/14 - Inclusão do tipo Básico TPlaca // v2.0-->

-<!-- PL_006u - 06/05/14 - Alterações Fuso-Horario // v2.0-->

-<!-- PL_006h - 13/05/11 - correções da NT 2011/004  // v2.0-->

-<!-- PL_006f - 29/05/10 - correcao do tipo TDec_1504 para limitar a quantidade de decimais para 4  // v2.0-->

-<!-- PL_006f - 09/05/10 - eliminação da possibilidade informar a Inscrição produtor rural na IEDest  // v2.0-->

-<!-- PL_006d - 04/10/09 - alterada a ordem do pattern do TIE - adequacao libxml  // v2.0-->

-<!-- PL_006d - 20/08/09 - acrescentado o tipo númerico com 10 casas decimais,15 casas inteiras e hora  // v2.0-->

-<!-- PL_005d - 11/08/09 - alteração no enumeration do tpais para nova tabela de paises do BACEN-->

-<!-- PL_005b - 24/10/08 - acrescentado a tabela do tpais   e outras alterações para eliminar os brancos no início e fim do campo   -->

 <xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" targetNamespace="http://www.portalfiscal.inf.br/nfe" elementFormDefault="qualified" attributeFormDefault="unqualified">

 	<xs:simpleType name="TCodUfIBGE">

@@ -45,4 +36,13 @@
 		</xs:restriction>

 	</xs:simpleType>

+	<xs:simpleType name="TnItem">

+		<xs:annotation>

+			<xs:documentation>Tipo correspondente ao atributo “nItem”</xs:documentation>

+		</xs:annotation>

+		<xs:restriction base="xs:string">

+			<xs:whiteSpace value="preserve"/>

+			<xs:pattern value="[1-9]{1}[0-9]{0,1}|[1-8]{1}[0-9]{2}|[9]{1}[0-8]{1}[0-9]{1}|[9]{1}[9]{1}[0]{1}"/>

+		</xs:restriction>

+	</xs:simpleType>

 	<xs:simpleType name="TCodMunIBGE">

 		<xs:annotation>

@@ -60,5 +60,5 @@
 		<xs:restriction base="xs:string">

 			<xs:whiteSpace value="preserve"/>

-      <xs:pattern value="(1[1-7]|2[1-9]|3[1,2,3,5]|4[1-3]|5[0-3])(0[6-9]|[1-9][\d])(0[1-9]|1[0-2])([\d]{14})([\d]{5})([\d]{9})([\d]{10})"/>

+			<xs:pattern value="[0-9]{6}[0-9A-Z]{12}[0-9]{26}"/>

 		</xs:restriction>

 	</xs:simpleType>

@@ -69,16 +69,16 @@
 		<xs:restriction base="xs:string">

 			<xs:whiteSpace value="preserve"/>

+			<xs:pattern value="[0-9]{15}|[0-9]{17}"/>

+		</xs:restriction>

+	</xs:simpleType>

+	<xs:simpleType name="TRec">

+		<xs:annotation>

+			<xs:documentation>Tipo Número do Recibo do envio de lote de NF-e</xs:documentation>

+		</xs:annotation>

+		<xs:restriction base="xs:string">

+			<xs:whiteSpace value="preserve"/>

 			<xs:pattern value="[0-9]{15}"/>

 		</xs:restriction>

 	</xs:simpleType>

-	<xs:simpleType name="TRec">

-		<xs:annotation>

-			<xs:documentation>Tipo Número do Recibo do envio de lote de NF-e</xs:documentation>

-		</xs:annotation>

-		<xs:restriction base="xs:string">

-			<xs:whiteSpace value="preserve"/>

-			<xs:pattern value="[0-9]{15}"/>

-		</xs:restriction>

-	</xs:simpleType>

 	<xs:simpleType name="TStat">

 		<xs:annotation>

@@ -87,5 +87,6 @@
 		<xs:restriction base="xs:string">

 			<xs:whiteSpace value="preserve"/>

-			<xs:pattern value="[0-9]{3}"/>

+			<xs:maxLength value="4"/>

+			<xs:pattern value="[0-9]{3,4}"/>

 		</xs:restriction>

 	</xs:simpleType>

@@ -96,14 +97,14 @@
 		<xs:restriction base="xs:string">

 			<xs:whiteSpace value="preserve"/>

-			<xs:pattern value="[0-9]{14}"/>

+			<xs:pattern value="[0-9A-Z]{12}[0-9]{2}"/>

 		</xs:restriction>

 	</xs:simpleType>

 	<xs:simpleType name="TCnpjVar">

 		<xs:annotation>

-			<xs:documentation>Tipo Número do CNPJ tmanho varíavel (3-14)</xs:documentation>

-		</xs:annotation>

-		<xs:restriction base="xs:string">

-			<xs:whiteSpace value="preserve"/>

-			<xs:pattern value="[0-9]{3,14}"/>

+			<xs:documentation>Tipo Número do CNPJ</xs:documentation>

+		</xs:annotation>

+		<xs:restriction base="xs:string">

+			<xs:whiteSpace value="preserve"/>

+			<xs:pattern value="[0-9A-Z]{12}[0-9]{2}"/>

 		</xs:restriction>

 	</xs:simpleType>

@@ -115,5 +116,5 @@
 			<xs:whiteSpace value="preserve"/>

 			<xs:maxLength value="14"/>

-			<xs:pattern value="[0-9]{0}|[0-9]{14}"/>

+			<xs:pattern value="[0-9]{0}|[0-9A-Z]{12}[0-9]{2}"/>

 		</xs:restriction>

 	</xs:simpleType>

@@ -808,13 +809,31 @@
 		</xs:restriction>

 	</xs:simpleType>

+	<xs:simpleType name="TDec_0302_04">

+		<xs:annotation>

+			<xs:documentation>Tipo Decimal com até 3 dígitos inteiros, podendo ter de 2 até 4 decimais</xs:documentation>

+		</xs:annotation>

... (truncated, 122 more lines — full diff in script output)
```

### `xmldsig-core-schema_v1.01.xsd` — diff

```diff
--- old/xmldsig-core-schema_v1.01.xsd
+++ new/xmldsig-core-schema_v1.01.xsd
@@ -96,3 +96,3 @@
 		</restriction>

 	</simpleType>

-</schema>

+</schema>
```

---

## 3. tiposBasico_v1.03.xsd — type-level diff

Types in OLD: 50
Types in NEW: 56
Types in both: 50  (changed: 8, unchanged: 42)

### Types only in NEW (added)
- `TDec1302`
- `TDec_0302_04`
- `TLatitude`
- `TLongitude`
- `TRSAKeyValueType`
- `TnItem`

### Types CHANGED (same name, different definition)
- `TCOrgaoIBGE`

```xml
<!-- OLD -->
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TCOrgaoIBGE">
		<xs:annotation>
			<xs:documentation>Tipo C&#243;digo de org&#227;o (UF da tabela do IBGE + 90 RFB)</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
			<xs:enumeration value="11"/>
			<xs:enumeration value="12"/>
			<xs:enumeration value="13"/>
			<xs:enumeration value="14"/>
			<xs:enumeration value="15"/>
			<xs:enumeration value="16"/>
			<xs:enumeration value="17"/>
			<xs:enumeration value="21"/>
			<xs:enumeration value="22"/>
			<xs:enumeration value="23"/>
			<xs:enumeration value="24"/>
			<xs:enumeration value="25"/>
			<xs:enumeration value="26"/>
			<xs:enumeration value="27"/>
			<xs:enumeration value="28"/>
			<xs:enumeration value="29"/>
			<xs:enumeration value="31"/>
			<xs:enumeration value="32"/>
			<xs:enumeration value="33"/>
			<xs:enumeration value="35"/>
			<xs:enumeration value="41"/>
			<xs:enumeration value="42"/>
			<xs:enumeration value="43"/>
			<xs:enumeration value="50"/>
			<xs:enumeration value="51"/>
			<xs:enumeration value="52"/>
			<xs:enumeration value="53"/>
			<xs:enumeration value="90"/>
			<xs:enumeration value="91"/>
			<xs:enumeration value="92"/>
		</xs:restriction>
	</xs:simpleType>


<!-- NEW -->
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TCOrgaoIBGE">
		<xs:annotation>
			<xs:documentation>Tipo C&#243;digo de org&#227;o (UF da tabela do IBGE + 90 RFB)</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
			<xs:enumeration value="11"/>
			<xs:enumeration value="12"/>
			<xs:enumeration value="13"/>
			<xs:enumeration value="14"/>
			<xs:enumeration value="15"/>
			<xs:enumeration value="16"/>
			<xs:enumeration value="17"/>
			<xs:enumeration value="21"/>
			<xs:enumeration value="22"/>
			<xs:enumeration value="23"/>
			<xs:enumeration value="24"/>
			<xs:enumeration value="25"/>
			<xs:enumeration value="26"/>
			<xs:enumeration value="27"/>
			<xs:enumeration value="28"/>
			<xs:enumeration value="29"/>
			<xs:enumeration value="31"/>
			<xs:enumeration value="32"/>
			<xs:enumeration value="33"/>
			<xs:enumeration value="35"/>
			<xs:enumeration value="41"/>
			<xs:enumeration value="42"/>
			<xs:enumeration value="43"/>
			<xs:enumeration value="50"/>
			<xs:enumeration value="51"/>
			<xs:enumeration value="52"/>
			<xs:enumeration value="53"/>
			<xs:enumeration value="90"/>
			<xs:enumeration value="91"/>
			<xs:enumeration value="92"/>
		</xs:restriction>
	</xs:simpleType>
	
```

- `TChNFe`

```xml
<!-- OLD -->
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TChNFe">
		<xs:annotation>
			<xs:documentation>Tipo Chave da Nota Fiscal Eletr&#244;nica</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
      <xs:pattern value="(1[1-7]|2[1-9]|3[1,2,3,5]|4[1-3]|5[0-3])(0[6-9]|[1-9][\d])(0[1-9]|1[0-2])([\d]{14})([\d]{5})([\d]{9})([\d]{10})"/>
		</xs:restriction>
	</xs:simpleType>
	

<!-- NEW -->
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TChNFe">
		<xs:annotation>
			<xs:documentation>Tipo Chave da Nota Fiscal Eletr&#244;nica</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
			<xs:pattern value="[0-9]{6}[0-9A-Z]{12}[0-9]{26}"/>
		</xs:restriction>
	</xs:simpleType>
	
```

- `TCnpj`

```xml
<!-- OLD -->
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TCnpj">
		<xs:annotation>
			<xs:documentation>Tipo N&#250;mero do CNPJ</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
			<xs:pattern value="[0-9]{14}"/>
		</xs:restriction>
	</xs:simpleType>
	

<!-- NEW -->
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TCnpj">
		<xs:annotation>
			<xs:documentation>Tipo N&#250;mero do CNPJ</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
			<xs:pattern value="[0-9A-Z]{12}[0-9]{2}"/>
		</xs:restriction>
	</xs:simpleType>
	
```

- `TCnpjOpc`

```xml
<!-- OLD -->
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TCnpjOpc">
		<xs:annotation>
			<xs:documentation>Tipo N&#250;mero do CNPJ Opcional</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
			<xs:maxLength value="14"/>
			<xs:pattern value="[0-9]{0}|[0-9]{14}"/>
		</xs:restriction>
	</xs:simpleType>
	

<!-- NEW -->
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TCnpjOpc">
		<xs:annotation>
			<xs:documentation>Tipo N&#250;mero do CNPJ Opcional</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
			<xs:maxLength value="14"/>
			<xs:pattern value="[0-9]{0}|[0-9A-Z]{12}[0-9]{2}"/>
		</xs:restriction>
	</xs:simpleType>
	
```

- `TCnpjVar`

```xml
<!-- OLD -->
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TCnpjVar">
		<xs:annotation>
			<xs:documentation>Tipo N&#250;mero do CNPJ tmanho var&#237;avel (3-14)</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
			<xs:pattern value="[0-9]{3,14}"/>
		</xs:restriction>
	</xs:simpleType>
	

<!-- NEW -->
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TCnpjVar">
		<xs:annotation>
			<xs:documentation>Tipo N&#250;mero do CNPJ</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
			<xs:pattern value="[0-9A-Z]{12}[0-9]{2}"/>
		</xs:restriction>
	</xs:simpleType>
	
```

- `TDec_1104Neg`

```xml
<!-- OLD -->
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_1104Neg">
    <xs:annotation>
      <xs:documentation>Tipo Decimal com at&#233; 15 d&#237;gitos, sendo 11 de corpo e at&#233; 4 decimais, aceitando valores negativos</xs:documentation>
    </xs:annotation>
    <xs:restriction base="xs:string">
      <xs:whiteSpace value="preserve"/>
      <xs:pattern value="0|0\.[0-9]{1,4}|[1-9]{1}[0-9]{0,10}|[1-9]{1}[0-9]{0,10}(\.[0-9]{1,4})?|-0\.[0-9]{1,4}|-[1-9]{1}[0-9]{0,10}|-[1-9]{1}[0-9]{0,10}(\.[0-9]{1,4})?"/>
    </xs:restriction>
  </xs:simpleType>
	

<!-- NEW -->
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_1104Neg">
		<xs:annotation>
			<xs:documentation>Tipo Decimal com at&#233; 15 d&#237;gitos, sendo 11 de corpo e at&#233; 4 decimais, aceitando valores negativos</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
			<xs:pattern value="0|0\.[0-9]{1,4}|[1-9]{1}[0-9]{0,10}|[1-9]{1}[0-9]{0,10}(\.[0-9]{1,4})?|-0\.[0-9]{1,4}|-[1-9]{1}[0-9]{0,10}|-[1-9]{1}[0-9]{0,10}(\.[0-9]{1,4})?"/>
		</xs:restriction>
	</xs:simpleType>
	
```

- `TProt`

```xml
<!-- OLD -->
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TProt">
		<xs:annotation>
			<xs:documentation>Tipo N&#250;mero do Protocolo de Status</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
			<xs:pattern value="[0-9]{15}"/>
		</xs:restriction>
	</xs:simpleType>
	

<!-- NEW -->
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TProt">
		<xs:annotation>
			<xs:documentation>Tipo N&#250;mero do Protocolo de Status</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
			<xs:pattern value="[0-9]{15}|[0-9]{17}"/>
		</xs:restriction>
	</xs:simpleType>
	
```

- `TStat`

```xml
<!-- OLD -->
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TStat">
		<xs:annotation>
			<xs:documentation>Tipo C&#243;digo da Mensagem enviada</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
			<xs:pattern value="[0-9]{3}"/>
		</xs:restriction>
	</xs:simpleType>
	

<!-- NEW -->
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TStat">
		<xs:annotation>
			<xs:documentation>Tipo C&#243;digo da Mensagem enviada</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
			<xs:maxLength value="4"/>
			<xs:pattern value="[0-9]{3,4}"/>
		</xs:restriction>
	</xs:simpleType>
	
```

---

## 4. tiposBasico_v4.00.xsd — new file (only in PL_010d_v1.03)

`tiposBasico_v4.00.xsd` contains 54 types.

### Type listing

- `TAmb` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TAmb"> <xs:annotation> <x...`
- `TCOrgaoIBGE` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TCOrgaoIBGE"> <xs:annotat...`
- `TChNFe` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TChNFe"> <xs:annotation> ...`
- `TCnpj` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TCnpj"> <xs:annotation> <...`
- `TCnpjOpc` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TCnpjOpc"> <xs:annotation...`
- `TCnpjVar` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TCnpjVar"> <xs:annotation...`
- `TCodMunIBGE` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TCodMunIBGE"> <xs:annotat...`
- `TCodUfIBGE` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TCodUfIBGE"> <xs:annotati...`
- `TCpf` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TCpf"> <xs:annotation> <x...`
- `TCpfVar` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TCpfVar"> <xs:annotation>...`
- `TData` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TData"> <xs:annotation> <...`
- `TDateTimeUTC` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDateTimeUTC"> <xs:annota...`
- `TDec_0104v` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_0104v"> <xs:annotati...`
- `TDec_0204v` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_0204v"> <xs:annotati...`
- `TDec_0302Max100` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_0302Max100"> <xs:ann...`
- `TDec_0302a04` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_0302a04"> <xs:annota...`
- `TDec_0302a04Max100` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_0302a04Max100"> <xs:...`
- `TDec_0302a04Opc` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_0302a04Opc"> <xs:ann...`
- `TDec_0304Max100` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_0304Max100"> <xs:ann...`
- `TDec_03v00a04Max100Opc` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_03v00a04Max100Opc"> ...`
- `TDec_0803v` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_0803v"> <xs:annotati...`
- `TDec_1104` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_1104"> <xs:annotatio...`
- `TDec_1104Opc` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_1104Opc"> <xs:annota...`
- `TDec_1104v` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_1104v"> <xs:annotati...`
- `TDec_1110v` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_1110v"> <xs:annotati...`
- `TDec_1203` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_1203"> <xs:annotatio...`
- `TDec_1204` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_1204"> <xs:annotatio...`
- `TDec_1204Opc` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_1204Opc"> <xs:annota...`
- `TDec_1204temperatura` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_1204temperatura"> <x...`
- `TDec_1204v` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_1204v"> <xs:annotati...`
- `TDec_1302` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_1302"> <xs:annotatio...`
- `TDec_1302Opc` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TDec_1302Opc"> <xs:annota...`
- `TIe` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TIe"> <xs:annotation> <xs...`
- `TIeDest` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TIeDest"> <xs:annotation>...`
- `TIeDestNaoIsento` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TIeDestNaoIsento"> <xs:an...`
- `TIeST` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TIeST"> <xs:annotation> <...`
- `TJust` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TJust"> <xs:annotation> <...`
- `TMed` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TMed"> <xs:annotation> <x...`
- `TMod` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TMod"> <xs:annotation> <x...`
- `TMotivo` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TMotivo"> <xs:annotation>...`
- `TNF` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TNF"> <xs:annotation> <xs...`
- `TPlaca` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TPlaca"> <xs:restriction ...`
- `TProt` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TProt"> <xs:annotation> <...`
- `TRSAKeyValueType` — `<xs:complexType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TRSAKeyValueType"> <xs:a...`
- `TRec` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TRec"> <xs:annotation> <x...`
- `TSerie` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TSerie"> <xs:annotation> ...`
- `TServ` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TServ"> <xs:annotation> <...`
- `TStat` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TStat"> <xs:annotation> <...`
- `TString` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TString"> <xs:annotation>...`
- `TTime` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TTime"> <xs:annotation> <...`
- `TUf` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TUf"> <xs:annotation> <xs...`
- `TUfEmi` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TUfEmi"> <xs:annotation> ...`
- `TVerAplic` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TVerAplic"> <xs:annotatio...`
- `Tano` — `<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="Tano"> <xs:annotation> <x...`

### Key type definitions

#### `TCnpj`

```xml
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TCnpj">
		<xs:annotation>
			<xs:documentation>Tipo N&#250;mero do CNPJ</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
			<xs:maxLength value="14"/>
			<xs:pattern value="[0-9A-Z]{12}[0-9]{2}"/>
		</xs:restriction>
	</xs:simpleType>
	
```

#### `TCpf`

```xml
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TCpf">
		<xs:annotation>
			<xs:documentation>Tipo N&#250;mero do CPF</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
			<xs:maxLength value="11"/>
			<xs:pattern value="[0-9]{11}"/>
		</xs:restriction>
	</xs:simpleType>
	
```

#### `TChNFe`

```xml
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TChNFe">
		<xs:annotation>
			<xs:documentation>Tipo Chave da Nota Fiscal Eletr&#244;nica</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
			<xs:maxLength value="44"/>
			<xs:pattern value="[0-9]{6}[0-9A-Z]{12}[0-9]{26}"/>
		</xs:restriction>
	</xs:simpleType>
	
```

#### `TJust`

```xml
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TJust">
		<xs:annotation>
			<xs:documentation>Tipo Justificativa</xs:documentation>
		</xs:annotation>
		<xs:restriction base="nfe:TString">
			<xs:minLength value="15"/>
			<xs:maxLength value="255"/>
		</xs:restriction>
	</xs:simpleType>
	
```

#### `TString`

```xml
<xs:simpleType xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:nfe="http://www.portalfiscal.inf.br/nfe" name="TString">
		<xs:annotation>
			<xs:documentation> Tipo string gen&#233;rico</xs:documentation>
		</xs:annotation>
		<xs:restriction base="xs:string">
			<xs:whiteSpace value="preserve"/>
			<xs:pattern value="[!-&#255;]{1}[ -&#255;]{0,}[!-&#255;]{1}|[!-&#255;]{1}"/>
		</xs:restriction>
	</xs:simpleType>
	
```

---

## 5. `TVerEvento` and `TVerEnvEvento` — WHERE they live

These two types define the **outer attribute** versions on
`envEvento/@versao`, `evento/@versao`, `retEvento/@versao`, `procEvento/@versao`.

**Important correction to the early analysis**: they were NOT removed. They moved:

- **OLD PL v1.01**: defined in `leiauteEventoCancNFe_v1.00.xsd` (lines 340-356)
  both with `pattern value="1\.00"`.
- **NEW PL_010d_v1.03**: defined in `leiauteEvento_v1.00.xsd` (lines 351-368)
  both with `pattern value="1\.00"`.

So the outer `versao="1.00"` is STILL required by the schema in both PLs.

### NEW `TVerEnvEvento` (PL_010d_v1.03 / `leiauteEvento_v1.00.xsd:351-358`)

```xml
<xs:simpleType name="TVerEnvEvento">
  <xs:annotation>
    <xs:documentation>Tipo Versão do EnvEvento</xs:documentation>
  </xs:annotation>
  <xs:restriction base="xs:string">
    <xs:whiteSpace value="preserve"/>
    <xs:pattern value="1\.00"/>
  </xs:restriction>
</xs:simpleType>
```

### NEW `TVerEvento` (PL_010d_v1.03 / `leiauteEvento_v1.00.xsd:360-368`)

```xml
<xs:simpleType name="TVerEvento">
  <xs:annotation>
    <xs:documentation>Tipo Versão do Evento</xs:documentation>
  </xs:annotation>
  <xs:restriction base="xs:string">
    <xs:whiteSpace value="preserve"/>
    <xs:pattern value="1\.00"/>
  </xs:restriction>
</xs:simpleType>
```

Note: the **inner `<verEvento>` element** (lines 72-81 of new schema) IS relaxed to an empty restriction — see section 6.

---

## 6. Cancellation-event schemas removed (old-only)

These files existed in `docs/Evento_Canc_PL_v1.01/` but are absent in PL_010d_v1.03.
The cancellation layout is now part of the generic `leiauteEvento_v1.00.xsd`,
where `detEvento` is declared as `xs:any` — schema-level enforcement of
`descEvento / nProt / xJust` has been removed.

- `e110111_v1.00.xsd`
- `envEventoCancNFe_v1.00.xsd`
- `eventoCancNFe_v1.00.xsd`
- `leiauteEventoCancNFe_v1.00.xsd`
- `procEventoCancNFe_v1.00.xsd`
- `retEnvEventoCancNFe_v1.00.xsd`

### Generic `leiauteEvento_v1.00.xsd` — `detEvento` element (new)

```xml
<xs:element xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:ds="http://www.w3.org/2000/09/xmldsig#" xmlns="http://www.portalfiscal.inf.br/nfe" name="detEvento">
							<xs:complexType>
								<xs:sequence>
									<xs:any processContents="skip" maxOccurs="unbounded">
										<xs:annotation>
											<xs:documentation>informa&#231;&#245;es espec&#237;ficas do evento</xs:documentation>
										</xs:annotation>
									</xs:any>
								</xs:sequence>
								<xs:anyAttribute processContents="skip"/>
							</xs:complexType>
						</xs:element>
						

```

### Inner `verEvento` element (new — empty restriction, any value accepted)

```xml
<xs:element xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:ds="http://www.w3.org/2000/09/xmldsig#" xmlns="http://www.portalfiscal.inf.br/nfe" name="verEvento">
							<xs:annotation>
								<xs:documentation>Vers&#227;o do Tipo do Evento</xs:documentation>
							</xs:annotation>
							<xs:simpleType>
								<xs:restriction base="xs:string">
									<xs:whiteSpace value="preserve"/>
								</xs:restriction>
							</xs:simpleType>
						</xs:element>
						

```

---

## 7. OLD `leiauteEventoCancNFe_v1.00.xsd` — `detEvento` and `verEvento` elements

### OLD `detEvento`

```xml
<xs:element xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:ds="http://www.w3.org/2000/09/xmldsig#" xmlns="http://www.portalfiscal.inf.br/nfe" name="detEvento">
							<xs:annotation>
								<xs:documentation>Schema XML de valida&#231;&#227;o do evento do cancelamento 1101111</xs:documentation>
							</xs:annotation>
							<xs:complexType>
								<xs:sequence>
									<xs:element name="descEvento">
										<xs:annotation>
											<xs:documentation>Descri&#231;&#227;o do Evento - &#8220;Cancelamento&#8221;</xs:documentation>
										</xs:annotation>
										<xs:simpleType>
											<xs:restriction base="xs:string">
												<xs:whiteSpace value="preserve"/>
												<xs:enumeration value="Cancelamento"/>
											</xs:restriction>
										</xs:simpleType>
									</xs:element>
									<xs:element name="nProt" type="TProt">
										<xs:annotation>
											<xs:documentation>N&#250;mero do Protocolo de Status da NF-e. 1 posi&#231;&#227;o (1 &#8211; Secretaria de Fazenda Estadual 2 &#8211; Receita Federal); 2 posi&#231;&#245;es ano; 10 seq&#252;encial no ano.</xs:documentation>
										</xs:annotation>
									</xs:element>
									<xs:element name="xJust" type="TJust">
										<xs:annotation>
											<xs:documentation>Justificativa do cancelamento</xs:documentation>
										</xs:annotation>
									</xs:element>
								</xs:sequence>
								<xs:attribute name="versao" use="required">
									<xs:simpleType>
										<xs:restriction base="xs:string">
											<xs:whiteSpace value="preserve"/>
											<xs:enumeration value="1.00"/>
										</xs:restriction>
									</xs:simpleType>
								</xs:attribute>
							</xs:complexType>
						</xs:element>
					

```

### OLD inner `verEvento` element — enumeration restricted to 1.00

```xml
<xs:element xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:ds="http://www.w3.org/2000/09/xmldsig#" xmlns="http://www.portalfiscal.inf.br/nfe" name="verEvento">
							<xs:annotation>
								<xs:documentation>Vers&#227;o do Tipo do Evento</xs:documentation>
							</xs:annotation>
							<xs:simpleType>
								<xs:restriction base="xs:string">
									<xs:whiteSpace value="preserve"/>
									<xs:enumeration value="1.00"/>
								</xs:restriction>
							</xs:simpleType>
						</xs:element>
						

```

---

## 8. Validation of real cancellation XML against new PL_010d_v1.03

We validate the **real PyNFe signed envelope** (`~/cancel_nfe_investigation/opencode/pynfe_env1_signed_event.xml`, wrapped in `<envEvento>`) against the new schema. This is the actual XML we send to SEFAZ MT homolog.

### NEW `envEvento_v1.00.xsd` — real PyNFe signed envelope

**Result**: PASS

```
VALID
```

### OLD `envEventoCancNFe_v1.00.xsd` — real PyNFe signed envelope (regression)

**Result**: PASS

```
VALID
```

---

## 9. Edge case: `versao="2.00"` on outer attributes (TVerEnvEvento / TVerEvento)

Mutating the real signed envelope to set `versao="2.00"` on envEvento, evento, and the inner `verEvento`. We expect both schemas to reject the outer attributes.

### NEW schema with versao=2.00

**Result**: FAIL (expected)

```
INVALID:
  - line 2: Element '{http://www.portalfiscal.inf.br/nfe}envEvento', attribute 'versao': [facet 'pattern'] The value '2.00' is not accepted by the pattern '1\.00'. (in /tmp/opencode/cancellation_v2_wrapped.xml)
  - line 4: Element '{http://www.portalfiscal.inf.br/nfe}evento', attribute 'versao': [facet 'pattern'] The value '2.00' is not accepted by the pattern '1\.00'. (in /tmp/opencode/cancellation_v2_wrapped.xml)
```

### OLD schema with versao=2.00

**Result**: FAIL (expected)

```
INVALID:
  - line 2: Element '{http://www.portalfiscal.inf.br/nfe}envEvento', attribute 'versao': [facet 'pattern'] The value '2.00' is not accepted by the pattern '1\.00'. (in /tmp/opencode/cancellation_v2_wrapped.xml)
  - line 4: Element '{http://www.portalfiscal.inf.br/nfe}evento', attribute 'versao': [facet 'pattern'] The value '2.00' is not accepted by the pattern '1\.00'. (in /tmp/opencode/cancellation_v2_wrapped.xml)
  - line 13: Element '{http://www.portalfiscal.inf.br/nfe}verEvento': [facet 'enumeration'] The value '2.00' is not an element of the set {'1.00'}. (in /tmp/opencode/cancellation_v2_wrapped.xml)
```

---

## 10. Edge case: outer `versao="1.00"`, inner `<verEvento>2.00</verEvento>`

Mutating only the inner `verEvento` to `2.00` while keeping outer `versao="1.00"`. The NEW schema should **pass** (inner restriction is empty); the OLD schema should **fail** (enumeration).

### NEW schema — outer=1.00 + inner verEvento=2.00

**Result**: PASS

```
VALID
```

### OLD schema — outer=1.00 + inner verEvento=2.00

**Result**: FAIL (expected)

```
INVALID:
  - line 13: Element '{http://www.portalfiscal.inf.br/nfe}verEvento': [facet 'enumeration'] The value '2.00' is not an element of the set {'1.00'}. (in /tmp/opencode/cancellation_inner_v2_wrapped.xml)
```

---

## 11. Pattern checks against `tiposBasico_v4.00.xsd`

`leiauteEvento_v1.00.xsd` imports `tiposBasico_v1.03.xsd` (not `v4.00`). PL_010d_v1.03 ships `tiposBasico_v4.00.xsd` for NF-e schemas (in `NFe/`). The cancellation event XML's regex fields (CNPJ/CPF/chNFe) are also valid against `v4.00`:

```
  CPF=83503463100                                         pattern=[0-9]{11}                                 -> PASS
  chNFe=51260800083503463100559200000000151980756634        pattern=[0-9]{6}[0-9A-Z]{12}[0-9]{26}             -> PASS
  infEvento Id=ID110111512608000835034631005592000000001519807566  pattern=ID[0-9]{12}[0-9A-Z]{12}[0-9]{28}          -> PASS
```

---

## 12. Summary

- The cancellation-event-specific schemas were **retired** (`leiauteEventoCancNFe_*`, `e110111_*`, etc.) — the layout is now part of the generic `leiauteEvento_v1.00.xsd`.
- The new `leiauteEvento_v1.00.xsd` declares `TVerEvento`/`TVerEnvEvento` (lines 351-368) with `pattern value="1\.00"` — same restriction as before, just relocated.
- The inner `<verEvento>` element inside `<infEvento>` has been **relaxed** (empty restriction, any value accepted by the schema).
- `<detEvento>` is now `xs:any` — schema no longer enforces `descEvento/nProt/xJust`.
- CNPJ regexes allow alphanumeric root+branch (numeric DV); CPF regex is unchanged.
- Access-key regex `TChNFe` is now `[0-9]{6}[0-9A-Z]{12}[0-9]{26}` (was strict numeric).
- Numeric-only access keys and CPFs (our case) are still valid.
- Our current `VersaoEvento = "1.00"` (in `pkg/nfe/xml/event.go:17`) is **correct** — both the new and old schemas require the outer `envEvento@versao` and `evento@versao` to be `1.00`.
- **No code change is required.** The cert OID diagnosis from before remains the primary cause of the cStat 617 rejection.
