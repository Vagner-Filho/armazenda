# Pesquisa: Integração com Nota Fiscal (NF-e/NFS-e/CT-e)

> **Status:** Em pesquisa  
> **Data:** 15/05/2026  
> **Objetivo:** Avaliar viabilidade de integrar o Armazenda com o sistema oficial de Notas Fiscais Eletrônicas (SEFAZ)

---

## 1. Contexto do Armazenda

O Armazenda é um sistema de gerenciamento de silos/armazéns agrícolas que hoje controla:

- **Entradas**: Recebimento de grãos (milho, soja) de produtores rurais ou propriedades próprias
- **Saídas/Despachos**: Envio de grãos para compradores (pessoas jurídicas com CNPJ)
- **Pessoas**: Pessoas físicas (CPF — produtores) e pessoas jurídicas (CNPJ — compradores)
- **Pesos e Qualidade**: Peso bruto, tara, peso líquido, umidade, impureza, avariados
- **Taxas de Armazenagem**: Cálculo de taxa de serviço sobre peso líquido

A integração com Nota Fiscal seria principalmente para **registrar as vendas de grãos** (saídas do silo) junto às secretarias da fazenda estaduais (SEFAZ).

---

## 2. Tipos de Documento Fiscal Relevantes

| Tipo | Descrição | Relevância para o Armazenda |
|------|-----------|----------------------------|
| **NF-e** (Nota Fiscal Eletrônica) | Emissão de nota fiscal de produto (venda de grãos) | **Alta** — vendas de milho/soja para compradores |
| **CT-e** (Conhecimento de Transporte Eletrônico) | Documento de transporte de cargas | **Média** — acompanhamento de frete nas saídas |
| **NFS-e** (Nota Fiscal de Serviços Eletrônica) | Prestação de serviços (armazenagem) | **Baixa/Média** — possível para taxa de armazenagem como serviço |

---

## 3. Opção 1: Integração Direta com SEFAZ (DIY)

### O que é preciso

Não existe uma **API REST oficial** da SEFAZ. A integração é feita via **Web Services SOAP/XML**, e **cada um dos 27 estados (+DF) opera seu próprio ambiente**.

| Requisito | Detalhes |
|-----------|----------|
| **Certificado Digital** | A1 (arquivo `.pfx`) ou A3 (token de hardware). Renovado **anualmente**. |
| **Geração de XML** | Construir XML seguindo os esquemas (PLs) da SEFAZ, que mudam todo ano |
| **Assinatura Digital** | Todo XML deve ser assinado com o certificado digital usando algoritmos específicos de canonicalização |
| **Cadeia de Certificados** | Manter a cadeia de certificados da SEFAZ por estado, atualizada quando houver rotação |
| **Web Services SOAP** | Comunicação HTTPS com TLS mútuo. WSDLs diferentes por estado. Serviços incluem: autorização, consulta de recibo, consulta de protocolo, cancelamento, carta de correção, inutilização, status |
| **Processamento Assíncrono vs Síncrono** | Alguns estados processam de forma síncrona, outros assíncrona — o sistema precisa lidar com ambos |
| **Armazenamento** | Obrigação legal de guardar todos os XMLs autorizados por **no mínimo 5 anos** |
| **Ambientes** | Homologação (teste) e Produção, com URLs diferentes por estado |
| **Validação de Schema** | XMLs são validados contra XSDs antes do envio; rejeitados se inválidos |

### Bibliotecas Open Source Existentes

- **`wmixvideo/nfe`** (Java, 754 estrelas): A biblioteca open source mais madura para comunicação direta com SEFAZ. Demonstra a complexidade real da integração.
- **`webmaniabr/NFe-Go`** (Go, 8 estrelas): Na verdade usa a API REST da WebmaniaBR, **não** a SEFAZ diretamente.
- **`chapzin/parse-efd-fiscal`** (Go, 69 estrelas): Faz parsing de arquivos SPED/EFD fiscal, mas **não** emite notas fiscais.

### Veredicto para DIY

Muito alta complexidade, manutenção contínua (SEFAZ muda esquemas e URLs regularmente), e exige conhecimento significativo em legislação tributária brasileira.

---

## 4. Opção 2: Provedores de API REST (Terceiros)

Empresas que abstraem toda a complexidade da SEFAZ e expõem uma API REST limpa. Você envia JSON, eles lidam com XML, assinaturas, WSDLs estaduais, gestão de certificados, etc.

| Provedor | Observações |
|----------|-------------|
| **WebmaniaBR** | Tem SDK em Go (`NFe-Go`). Marca conhecida. Cobrança por nota emitida. |
| **Nuvem Fiscal** | API REST moderna. Boa experiência para desenvolvedores. |
| **Focus NFe (Simplifique)** | Popular entre ERPs. API REST. |
| **TecnoSpeed (PlugNotas)** | Oferece também NFS-e além de NF-e. |
| **Agilize** | Outra opção de API REST. |

### Comparativo Direto vs Terceirizado

| Aspecto | SEFAZ Direto | API de Terceiro |
|---------|--------------|-----------------|
| **Custo** | Só certificado (~R$ 100–300/ano) + tempo de desenvolvimento | Taxa por nota (~R$ 0,30–1,50/nota) ou plano mensal |
| **Manutenção** | Muito alta (atualizações de esquema, cadeia de certificados, mudanças de URL) | Muito baixa (provedor lida com isso) |
| **Tempo para Produção** | 3–6 meses para uma solução robusta | 2–4 semanas |
| **Risco de Falha** | Você precisa lidar com indisponibilidade da SEFAZ, retries, filas | Provedor gerencia resiliência |
| **Customização** | Controle total | Limitado à superfície da API do provedor |
| **Certificado** | Você gerencia a renovação | Provedor frequentemente gerencia ou hospeda |

---

## 5. Gaps no Armazenda Hoje

Antes de qualquer integração, o sistema precisa de campos de **preço e valor monetário**:

| Gap | Por que é necessário |
|-----|----------------------|
| **Preço unitário por kg** ou **valor total** no Despacho | NF-e exige valores monetários (preço unitário × quantidade = total) |
| **Códigos CFOP** | Código Fiscal de Operações e Prestações (ex.: 5102 para venda de grãos) |
| **Códigos NCM** | Nomenclatura Comum do Mercosul — milho (1005.90.00) e soja (1201.00.10) |
| **Configuração de ICMS/PIS/COFINS** | Alíquotas variam por estado (origem/destino) e produto |
| **Rastreamento de status da nota** | Pendente, autorizada, denegada, cancelada, corrigida |
| **Armazenamento do XML da nota** | Obrigação legal |
| **Geração de DANFE** | Representação em PDF da NF-e para acompanhamento de transporte |

---

## 6. Perguntas Pendentes

> Essas perguntas precisam ser respondidas antes de decidir o caminho a seguir.

1. **Qual tipo de documento fiscal queremos começar?** NF-e (venda de produto), CT-e (transporte) ou NFS-e (serviço de armazenagem)?

2. **Qual é o volume mensal de notas?** Quantas notas fiscais por mês o sistema precisaria emitir? Isso impacta diretamente a viabilidade de custo da opção terceirizada vs. própria.

3. **Existe um contador ou equipe fiscal** que possa fornecer os códigos CFOP, NCM e configuração de ICMS? Isso é crítico independente da abordagem técnica.

4. **Qual é a prioridade: velocidade ou custo?** Terceirizado coloca a feature em produção em semanas mas tem custo recorrente. Próprio economiza dinheiro a longo prazo, mas é um investimento de 3–6 meses com manutenção contínua.

5. **Já existe um certificado digital (A1) para o CNPJ da propriedade/fazenda?**

---

## 7. Recomendação Provisória

Dado o estágio atual do Armazenda e a complexidade da conformidade fiscal brasileira, a abordagem sugerida é **híbrida**:

- **Fase 1**: Integrar com uma **API REST de terceiro** (ex.: WebmaniaBR ou Nuvem Fiscal) para ter NF-e nos despachos funcionando rapidamente. Isso prova a feature e valida a necessidade de negócio.
- **Fase 2**: Quando o volume justificar (centenas/milhares de notas/mês), avaliar a migração para **integração direta com SEFAZ** usando uma biblioteca robusta ou construindo um comunicador próprio em Go. Até lá, você terá aprendido exatamente quais são suas necessidades.

---

## 8. Próximos Passos Possíveis

- [ ] Responder às perguntas pendentes da Seção 5
- [ ] Definir qual tipo de documento fiscal será o escopo inicial
- [ ] Modelar entidades e schema do banco para rastreamento de notas fiscais
- [ ] Criar um plano de integração detalhado (escolhendo o provedor ou a arquitetura DIY)
- [ ] Adicionar campos de valor monetário, CFOP e NCM nas entidades de Despacho e Pessoa
