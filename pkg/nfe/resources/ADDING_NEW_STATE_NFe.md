# How to Add a New State to the NF-e Integration

> **Document:** Adding a new Brazilian state to `pkg/nfe`  
> **Prerequisite:** The state must issue NF-e (modelo 55) via SEFAZ web services.

---

## Overview

Adding a new state requires changes to **two files**:
1. `pkg/nfe/sefaz/endpoints.go` — register SEFAZ web service endpoints
2. `model/armazenda_database/seed_municipios.go` — add the state's municipalities

No changes are needed to:
- XML builder (`pkg/nfe/xml/`)
- Digital signature (`pkg/nfe/sign/`)
- DANFE generator (`pkg/nfe/service/danfe.go`)
- Service orchestration (`pkg/nfe/service/invoice_service.go`)
- Database schema (`nfe_farm_config` already has `emitter_uf`)

---

## Step 1: Determine the State's Infrastructure Type

Brazilian states use one of three NF-e authorization infrastructures:

| Type | Description | URL Pattern |
|------|-------------|-------------|
| **Own** | State runs its own SEFAZ servers | `https://nfe.sefaz.{uf}.gov.br/...` or custom |
| **SVRS** | Shared Virtual Environment (RS-based) | `https://nfe.svrs.rs.gov.br/...` |
| **SVAN** | Shared Virtual Environment (AN-based, legacy) | `https://nfe.fazenda.{uf}.gov.br/...` |

Check the official SEFAZ portal to confirm:  
https://www.nfe.fazenda.gov.br/portal/webServices.aspx

---

## Step 2: Gather the 7 Web Service Endpoints

For the new state, collect the production and homologation URLs for all 7 services:

1. `NFeAutorizacao4` — Send NF-e for authorization
2. `NFeRetAutorizacao4` — Retrieve authorization result
3. `NfeConsulta4` — Query NF-e protocol/status
4. `NfeStatusServico4` — Check service availability
5. `RecepcaoEvento4` — Receive events (cancellation, correction letter)
6. `NfeInutilizacao4` — Number inutilization
7. `CadConsultaCadastro4` — Query taxpayer registry

Also collect the **SOAP namespace** for each service. It typically follows this pattern:
```
http://www.portalfiscal.inf.br/nfe/wsdl/{ServiceName}
```

Example for MT:
```
http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4
```

---

## Step 3: Create the EndpointSet

Open `pkg/nfe/sefaz/endpoints.go`.

### For a State with Own Infrastructure

Create a new `EndpointSet` variable below `MTEndpointSet`:

```go
// RSEndpointSet holds all NF-e web service endpoints for Rio Grande do Sul.
var RSEndpointSet = &EndpointSet{
    Name:      "RS",
    InfraType: InfraOwn,
    Endpoints: []Endpoint{
        {
            Name:         "NFeAutorizacao4",
            Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4",
            Production:   "https://nfe.sefaz.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx",
            Homologation: "https://nfe-homologacao.sefaz.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx",
        },
        {
            Name:         "NFeRetAutorizacao4",
            Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRetAutorizacao4",
            Production:   "https://nfe.sefaz.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx",
            Homologation: "https://nfe-homologacao.sefaz.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx",
        },
        // ... repeat for all 7 services
    },
}
```

### For an SVRS State (Shared Infrastructure)

If the state uses SVRS, **reuse** the shared `EndpointSet`. You only need to create the `EndpointSet` once for SVRS, then register multiple states to it:

```go
// SVRSEndpointSet holds all NF-e web service endpoints for SVRS virtual environment.
var SVRSEndpointSet = &EndpointSet{
    Name:      "SVRS",
    InfraType: InfraSVRS,
    Endpoints: []Endpoint{
        {
            Name:         "NFeAutorizacao4",
            Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4",
            Production:   "https://nfe.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx",
            Homologation: "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx",
        },
        // ... all 7 services
    },
}
```

---

## Step 4: Register the State

Find the `StateRegistry` map in the same file and add your state:

### Own Infrastructure

```go
var StateRegistry = map[string]*EndpointSet{
    "MT": MTEndpointSet,
    "RS": RSEndpointSet, // <-- added
}
```

### SVRS Infrastructure

```go
var StateRegistry = map[string]*EndpointSet{
    "MT": MTEndpointSet,
    "AC": SVRSEndpointSet, // <-- added
    "AL": SVRSEndpointSet, // <-- added
    "AP": SVRSEndpointSet, // <-- added
    "RJ": SVRSEndpointSet, // <-- added
}
```

---

## Step 5: Add the State's Municipalities

### Why This Is Required

The NF-e XML requires the emitter's and recipient's municipality codes (`cMun`) in the `enderEmit` and `enderDest` elements. The system resolves municipality **names** to IBGE **codes** via the `ibge_municipio` table. If the new state's municipalities are missing, the NF-e XML will have an empty `cMun` field and SEFAZ will reject it.

### How to Add Them

#### Option A: Update the Seed File (Recommended)

Open `model/armazenda_database/seed_municipios.go` and append the new state's municipalities to `municipiosSeedData`:

```go
var municipiosSeedData = []municipioSeed{
    // ... existing MT municipalities ...

    // Rio Grande do Sul (RS) - 497 municipalities
    {"4300034", "Aceguá", "RS"},
    {"4300059", "Água Santa", "RS"},
    // ... all remaining RS municipalities ...
}
```

**To fetch the complete list from IBGE:**

```bash
curl -s "https://servicodados.ibge.gov.br/api/v1/localidades/estados/RS/municipios" | \
  jq -r '.[] | "{\"\"\"\(.id)\"\"\", \"\(.nome)\"\"\", \"RS\"},"'
```

> Replace `RS` with the target state's UF code.

#### Option B: SQL Migration

If you prefer not to modify the seed file, create a migration:

```sql
-- model/armazenda_database/migrations/000XXX_add_rs_municipios.sql
INSERT INTO ibge_municipio (code, name, uf) VALUES
('4300034', 'Aceguá', 'RS'),
('4300059', 'Água Santa', 'RS'),
-- ... all municipalities ...
ON CONFLICT DO NOTHING;
```

#### Option C: Direct SQL (One-Off)

For quick testing, insert directly into the database:

```sql
INSERT INTO ibge_municipio (code, name, uf) VALUES ('4300034', 'Aceguá', 'RS');
```

> **Note:** The existing seed already includes all **state capitals** and **major cities** across Brazil. You only need to add the remaining municipalities for the new state.

---

## Step 6: Verify

Build the project to ensure no compilation errors:

```bash
go build ./...
```

Run tests:

```bash
go test ./...
```

If you updated the seed file, the municipalities will be inserted automatically on the next application startup (or when `seedMunicipios()` runs).

---

## Step 7: Configure the Farm

In the Armazenda UI, go to **Farm Settings → NF-e Configuration** and set:

- **Emitter UF**: the new state code (e.g., `RS`, `AC`)
- All other fields (certificate, tax regime, etc.) as usual

The system will automatically resolve the correct SEFAZ endpoints based on `emitter_uf`.

---

## Example: Complete Diff for Adding RS

### Endpoints

```diff
// pkg/nfe/sefaz/endpoints.go

+ var RSEndpointSet = &EndpointSet{
+     Name:      "RS",
+     InfraType: InfraOwn,
+     Endpoints: []Endpoint{
+         {
+             Name:         "NFeAutorizacao4",
+             Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4",
+             Production:   "https://nfe.sefaz.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx",
+             Homologation: "https://nfe-homologacao.sefaz.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx",
+         },
+         // ... 6 more services
+     },
+ }

  var StateRegistry = map[string]*EndpointSet{
      "MT": MTEndpointSet,
+     "RS": RSEndpointSet,
  }
```

### Municipalities

```diff
// model/armazenda_database/seed_municipios.go

  var municipiosSeedData = []municipioSeed{
      // ... existing entries ...

+     // Rio Grande do Sul (RS)
+     {"4300034", "Aceguá", "RS"},
+     {"4300059", "Água Santa", "RS"},
+     // ... (all 497 municipalities)
  }
```

That's it. Only `endpoints.go` and `seed_municipios.go` need to change.

---

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| `state X not registered in endpoint registry` | Missing entry in `StateRegistry` | Add the state UF to `StateRegistry` |
| `service Y not found for state X` | Typo in service name | Ensure the `Name` field matches exactly (case-sensitive) |
| Wrong URL returned for SVRS state | Using `InfraOwn` instead of `InfraSVRS` | Set `InfraType: InfraSVRS` in the `EndpointSet` |
| SOAP namespace mismatch | Wrong `Namespace` field | Verify the xmlns from the official WSDL |
| SEFAZ rejects with `cMun` error | Municipality code missing from `ibge_municipio` | Add the municipality to `seed_municipios.go` or insert into DB |
| `cMun` is empty in generated XML | `GetMunicipio()` returned no rows | Verify municipality name matches exactly (case-insensitive); add if missing |

---

## Current State Registry

As of the last update, the following states are registered for SEFAZ communication:

| State | Infrastructure | EndpointSet |
|-------|---------------|-------------|
| `MT` | Own | `MTEndpointSet` |

### Current Municipality Coverage

| State | Municipality Coverage | Source |
|-------|----------------------|--------|
| `MT` | **All 141 municipalities** | `seed_municipios.go` |
| All other states | **State capitals + major cities only** | `seed_municipios.go` |

> **When adding a new state, you must add all of its municipalities** (see Step 5). The existing seed only guarantees state capitals and major commercial cities for cross-state operations.

To see the full current registry, check `pkg/nfe/sefaz/endpoints.go`.
