package armazenda_database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type municipioSeed struct {
	Code string
	Name string
	UF   string
}

var municipiosSeedData = []municipioSeed{
	// Mato Grosso (MT) - 141 municipalities
	{"5100102", "Acorizal", "MT"}, {"5100201", "Água Boa", "MT"}, {"5100250", "Alta Floresta", "MT"},
	{"5100300", "Alto Araguaia", "MT"}, {"5100359", "Alto Boa Vista", "MT"}, {"5100409", "Alto Garças", "MT"},
	{"5100508", "Alto Paraguai", "MT"}, {"5100607", "Alto Taquari", "MT"}, {"5100805", "Apiacás", "MT"},
	{"5101001", "Araguaiana", "MT"}, {"5101209", "Araguainha", "MT"}, {"5101258", "Araputanga", "MT"},
	{"5101308", "Arenápolis", "MT"}, {"5101407", "Aripuanã", "MT"}, {"5101605", "Barão de Melgaço", "MT"},
	{"5101704", "Barra do Bugres", "MT"}, {"5101803", "Barra do Garças", "MT"}, {"5101837", "Boa Esperança do Norte", "MT"},
	{"5101852", "Bom Jesus do Araguaia", "MT"}, {"5101902", "Brasnorte", "MT"}, {"5102504", "Cáceres", "MT"},
	{"5102603", "Campinápolis", "MT"}, {"5102637", "Campo Novo do Parecis", "MT"}, {"5102678", "Campo Verde", "MT"},
	{"5102686", "Campos de Júlio", "MT"}, {"5102694", "Canabrava do Norte", "MT"}, {"5102702", "Canarana", "MT"},
	{"5102793", "Carlinda", "MT"}, {"5102850", "Castanheira", "MT"}, {"5103007", "Chapada dos Guimarães", "MT"},
	{"5103056", "Cláudia", "MT"}, {"5103106", "Cocalinho", "MT"}, {"5103205", "Colíder", "MT"},
	{"5103254", "Colniza", "MT"}, {"5103304", "Comodoro", "MT"}, {"5103353", "Confresa", "MT"},
	{"5103361", "Conquista D'Oeste", "MT"}, {"5103379", "Cotriguaçu", "MT"}, {"5103403", "Cuiabá", "MT"},
	{"5103437", "Curvelândia", "MT"}, {"5103452", "Denise", "MT"}, {"5103502", "Diamantino", "MT"},
	{"5103601", "Dom Aquino", "MT"}, {"5103700", "Feliz Natal", "MT"}, {"5103809", "Figueirópolis D'Oeste", "MT"},
	{"5103858", "Gaúcha do Norte", "MT"}, {"5103908", "General Carneiro", "MT"}, {"5103957", "Glória D'Oeste", "MT"},
	{"5104104", "Guarantã do Norte", "MT"}, {"5104203", "Guiratinga", "MT"}, {"5104500", "Indiavaí", "MT"},
	{"5104526", "Ipiranga do Norte", "MT"}, {"5104542", "Itanhangá", "MT"}, {"5104559", "Itaúba", "MT"},
	{"5104609", "Itiquira", "MT"}, {"5104807", "Jaciara", "MT"}, {"5104906", "Jangada", "MT"},
	{"5105002", "Jauru", "MT"}, {"5105101", "Juara", "MT"}, {"5105150", "Juína", "MT"},
	{"5105176", "Juruena", "MT"}, {"5105200", "Juscimeira", "MT"}, {"5105234", "Lambari D'Oeste", "MT"},
	{"5105259", "Lucas do Rio Verde", "MT"}, {"5105309", "Luciara", "MT"}, {"5105507", "Vila Bela da Santíssima Trindade", "MT"},
	{"5105580", "Marcelândia", "MT"}, {"5105606", "Matupá", "MT"}, {"5105622", "Mirassol d'Oeste", "MT"},
	{"5105903", "Nobres", "MT"}, {"5106000", "Nortelândia", "MT"}, {"5106109", "Nossa Senhora do Livramento", "MT"},
	{"5106158", "Nova Bandeirantes", "MT"}, {"5106174", "Nova Nazaré", "MT"}, {"5106182", "Nova Lacerda", "MT"},
	{"5106190", "Nova Santa Helena", "MT"}, {"5106208", "Nova Brasilândia", "MT"}, {"5106216", "Nova Canaã do Norte", "MT"},
	{"5106224", "Nova Mutum", "MT"}, {"5106232", "Nova Olímpia", "MT"}, {"5106240", "Nova Ubiratã", "MT"},
	{"5106257", "Nova Xavantina", "MT"}, {"5106265", "Novo Mundo", "MT"}, {"5106273", "Novo Horizonte do Norte", "MT"},
	{"5106281", "Novo São Joaquim", "MT"}, {"5106299", "Paranaíta", "MT"}, {"5106307", "Paranatinga", "MT"},
	{"5106315", "Novo Santo Antônio", "MT"}, {"5106372", "Pedra Preta", "MT"}, {"5106422", "Peixoto de Azevedo", "MT"},
	{"5106455", "Planalto da Serra", "MT"}, {"5106505", "Poconé", "MT"}, {"5106653", "Pontal do Araguaia", "MT"},
	{"5106703", "Ponte Branca", "MT"}, {"5106752", "Pontes e Lacerda", "MT"}, {"5106778", "Porto Alegre do Norte", "MT"},
	{"5106802", "Porto dos Gaúchos", "MT"}, {"5106828", "Porto Esperidião", "MT"}, {"5106851", "Porto Estrela", "MT"},
	{"5107008", "Poxoréu", "MT"}, {"5107040", "Primavera do Leste", "MT"}, {"5107065", "Querência", "MT"},
	{"5107107", "São José dos Quatro Marcos", "MT"}, {"5107156", "Reserva do Cabaçal", "MT"}, {"5107180", "Ribeirão Cascalheira", "MT"},
	{"5107198", "Ribeirãozinho", "MT"}, {"5107206", "Rio Branco", "MT"}, {"5107248", "Santa Carmem", "MT"},
	{"5107263", "Santo Afonso", "MT"}, {"5107297", "São José do Povo", "MT"}, {"5107305", "São José do Rio Claro", "MT"},
	{"5107354", "São José do Xingu", "MT"}, {"5107404", "São Pedro da Cipa", "MT"}, {"5107578", "Rondolândia", "MT"},
	{"5107602", "Rondonópolis", "MT"}, {"5107701", "Rosário Oeste", "MT"}, {"5107743", "Santa Cruz do Xingu", "MT"},
	{"5107750", "Salto do Céu", "MT"}, {"5107768", "Santa Rita do Trivelato", "MT"}, {"5107776", "Santa Terezinha", "MT"},
	{"5107792", "Santo Antônio do Leste", "MT"}, {"5107800", "Santo Antônio de Leverger", "MT"}, {"5107859", "São Félix do Araguaia", "MT"},
	{"5107875", "Sapezal", "MT"}, {"5107883", "Serra Nova Dourada", "MT"}, {"5107909", "Sinop", "MT"},
	{"5107925", "Sorriso", "MT"}, {"5107941", "Tabaporã", "MT"}, {"5107958", "Tangará da Serra", "MT"},
	{"5108006", "Tapurah", "MT"}, {"5108055", "Terra Nova do Norte", "MT"}, {"5108105", "Tesouro", "MT"},
	{"5108204", "Torixoréu", "MT"}, {"5108303", "União do Sul", "MT"}, {"5108352", "Vale de São Domingos", "MT"},
	{"5108402", "Várzea Grande", "MT"}, {"5108501", "Vera", "MT"}, {"5108600", "Vila Rica", "MT"},
	{"5108808", "Nova Guarita", "MT"}, {"5108857", "Nova Marilândia", "MT"}, {"5108907", "Nova Maringá", "MT"},
	{"5108956", "Nova Monte Verde", "MT"},

	// State capitals
	{"1100205", "Porto Velho", "RO"}, {"1200401", "Rio Branco", "AC"}, {"1302603", "Manaus", "AM"},
	{"1400100", "Boa Vista", "RR"}, {"1501402", "Belém", "PA"}, {"1600303", "Macapá", "AP"},
	{"1721000", "Palmas", "TO"}, {"2111300", "São Luís", "MA"}, {"2211001", "Teresina", "PI"},
	{"2304400", "Fortaleza", "CE"}, {"2408102", "Natal", "RN"}, {"2507507", "João Pessoa", "PB"},
	{"2611606", "Recife", "PE"}, {"2704302", "Maceió", "AL"}, {"2800308", "Aracaju", "SE"},
	{"2927408", "Salvador", "BA"}, {"3106200", "Belo Horizonte", "MG"}, {"3205309", "Vitória", "ES"},
	{"3304557", "Rio de Janeiro", "RJ"}, {"3550308", "São Paulo", "SP"}, {"4106902", "Curitiba", "PR"},
	{"4205407", "Florianópolis", "SC"}, {"4314902", "Porto Alegre", "RS"}, {"5002704", "Campo Grande", "MS"},
	{"5208707", "Goiânia", "GO"}, {"5300108", "Brasília", "DF"},

	// Major agricultural/commercial cities
	{"3509502", "Campinas", "SP"}, {"3518800", "Guarulhos", "SP"}, {"3543402", "Ribeirão Preto", "SP"},
	{"3549904", "São José do Rio Preto", "SP"}, {"4113700", "Londrina", "PR"}, {"4115200", "Maringá", "PR"},
	{"4104808", "Cascavel", "PR"}, {"4202404", "Blumenau", "SC"}, {"4208203", "Joinville", "SC"},
	{"4305108", "Caxias do Sul", "RS"}, {"4315602", "Rio Grande", "RS"}, {"3170206", "Uberlândia", "MG"},
	{"3106705", "Betim", "MG"}, {"2919207", "Feira de Santana", "BA"}, {"5008006", "Três Lagoas", "MS"},
}

func seedMunicipios(c *pgx.Conn) {
	var count int
	err := c.QueryRow(context.Background(), "SELECT COUNT(*) FROM ibge_municipio").Scan(&count)
	if err != nil {
		fmt.Printf("error counting ibge_municipio: %v\n", err.Error())
		return
	}
	if count > 0 {
		return
	}

	for _, m := range municipiosSeedData {
		_, insertErr := c.Exec(context.Background(),
			"INSERT INTO ibge_municipio (code, name, uf) VALUES ($1, $2, $3)",
			m.Code, m.Name, m.UF)
		if insertErr != nil {
			fmt.Printf("error inserting municipio %s: %v\n", m.Code, insertErr.Error())
		}
	}
}
