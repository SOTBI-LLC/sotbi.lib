package dadata

type Suggestion struct {
	Value             string `json:"value"`
	UnrestrictedValue string `json:"unrestricted_value"`
	Data              struct {
		Kpp        string `json:"kpp"`
		KppLargest any    `json:"kpp_largest"`
		Capital    struct {
			Type  string  `json:"type"`
			Value float64 `json:"value"`
		} `json:"capital"`
		Invalid    any `json:"invalid"`
		Management struct {
			Name         string `json:"name"`
			Post         string `json:"post"`
			StartDate    int64  `json:"start_date"`
			Disqualified any    `json:"disqualified"`
		} `json:"management"`
		Founders []struct {
			Ogrn  string `json:"ogrn"`
			Inn   string `json:"inn"`
			Name  string `json:"name"`
			Hid   string `json:"hid"`
			Type  string `json:"type"`
			Share struct {
				Value int    `json:"value"`
				Type  string `json:"type"`
			} `json:"share"`
			Invalidity any   `json:"invalidity"`
			StartDate  int64 `json:"start_date"`
		} `json:"founders"`
		Managers []struct {
			Inn string `json:"inn"`
			Fio struct {
				Surname    string `json:"surname"`
				Name       string `json:"name"`
				Patronymic string `json:"patronymic"`
				Gender     string `json:"gender"`
				Source     string `json:"source"`
				Qc         any    `json:"qc"`
			} `json:"fio"`
			Post       string `json:"post"`
			Hid        string `json:"hid"`
			Type       string `json:"type"`
			Invalidity any    `json:"invalidity"`
			StartDate  int64  `json:"start_date"`
		} `json:"managers"`
		Predecessors any    `json:"predecessors"`
		Successors   any    `json:"successors"`
		BranchType   string `json:"branch_type"`
		BranchCount  int    `json:"branch_count"`
		Source       any    `json:"source"`
		Qc           any    `json:"qc"`
		Hid          string `json:"hid"`
		Type         string `json:"type"`
		State        struct {
			Status           string `json:"status"`
			Code             any    `json:"code"`
			ActualityDate    int64  `json:"actuality_date"`
			RegistrationDate int64  `json:"registration_date"`
			LiquidationDate  any    `json:"liquidation_date"`
		} `json:"state"`
		Opf struct {
			Type  string `json:"type"`
			Code  string `json:"code"`
			Full  string `json:"full"`
			Short string `json:"short"`
		} `json:"opf"`
		Name struct {
			FullWithOpf  string `json:"full_with_opf"`
			ShortWithOpf string `json:"short_with_opf"`
			Latin        any    `json:"latin"`
			Full         string `json:"full"`
			Short        string `json:"short"`
		} `json:"name"`
		Inn    string `json:"inn"`
		Ogrn   string `json:"ogrn"`
		Okpo   string `json:"okpo"`
		Okato  string `json:"okato"`
		Oktmo  string `json:"oktmo"`
		Okogu  string `json:"okogu"`
		Okfs   string `json:"okfs"`
		Okved  string `json:"okved"`
		Okveds []struct {
			Main bool   `json:"main"`
			Type string `json:"type"`
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"okveds"`
		Authorities struct {
			FtsRegistration struct {
				Type    string `json:"type"`
				Code    string `json:"code"`
				Name    string `json:"name"`
				Address string `json:"address"`
			} `json:"fts_registration"`
			FtsReport struct {
				Type    string `json:"type"`
				Code    string `json:"code"`
				Name    string `json:"name"`
				Address any    `json:"address"`
			} `json:"fts_report"`
			Pf struct {
				Type    string `json:"type"`
				Code    string `json:"code"`
				Name    string `json:"name"`
				Address any    `json:"address"`
			} `json:"pf"`
			Sif struct {
				Type    string `json:"type"`
				Code    string `json:"code"`
				Name    string `json:"name"`
				Address any    `json:"address"`
			} `json:"sif"`
		} `json:"authorities"`
		Documents struct {
			FtsRegistration struct {
				Type           string `json:"type"`
				Series         string `json:"series"`
				Number         string `json:"number"`
				IssueDate      int64  `json:"issue_date"`
				IssueAuthority string `json:"issue_authority"`
			} `json:"fts_registration"`
			FtsReport struct {
				Type           string `json:"type"`
				Series         any    `json:"series"`
				Number         any    `json:"number"`
				IssueDate      int64  `json:"issue_date"`
				IssueAuthority string `json:"issue_authority"`
			} `json:"fts_report"`
			PfRegistration struct {
				Type           string `json:"type"`
				Series         any    `json:"series"`
				Number         string `json:"number"`
				IssueDate      int64  `json:"issue_date"`
				IssueAuthority string `json:"issue_authority"`
			} `json:"pf_registration"`
			SifRegistration struct {
				Type           string `json:"type"`
				Series         any    `json:"series"`
				Number         string `json:"number"`
				IssueDate      int64  `json:"issue_date"`
				IssueAuthority string `json:"issue_authority"`
			} `json:"sif_registration"`
			Smb struct {
				Category       string `json:"category"`
				Type           string `json:"type"`
				Series         any    `json:"series"`
				Number         any    `json:"number"`
				IssueDate      int64  `json:"issue_date"`
				IssueAuthority any    `json:"issue_authority"`
			} `json:"smb"`
		} `json:"documents"`
		Licenses any `json:"licenses"`
		Finance  struct {
			TaxSystem any `json:"tax_system"`
			Income    int `json:"income"`
			Expense   int `json:"expense"`
			Revenue   int `json:"revenue"`
			Debt      any `json:"debt"`
			Penalty   any `json:"penalty"`
			Year      int `json:"year"`
		} `json:"finance"`
		Address struct {
			Value             string `json:"value"`
			UnrestrictedValue string `json:"unrestricted_value"`
			Invalidity        any    `json:"invalidity"`
			Data              struct {
				PostalCode           string `json:"postal_code"`
				Country              string `json:"country"`
				CountryIsoCode       string `json:"country_iso_code"`
				FederalDistrict      string `json:"federal_district"`
				RegionFiasId         string `json:"region_fias_id"`
				RegionKladrId        string `json:"region_kladr_id"`
				RegionIsoCode        string `json:"region_iso_code"`
				RegionWithType       string `json:"region_with_type"`
				RegionType           string `json:"region_type"`
				RegionTypeFull       string `json:"region_type_full"`
				Region               string `json:"region"`
				AreaFiasId           any    `json:"area_fias_id"`
				AreaKladrId          any    `json:"area_kladr_id"`
				AreaWithType         any    `json:"area_with_type"`
				AreaType             any    `json:"area_type"`
				AreaTypeFull         any    `json:"area_type_full"`
				Area                 any    `json:"area"`
				CityFiasId           string `json:"city_fias_id"`
				CityKladrId          string `json:"city_kladr_id"`
				CityWithType         string `json:"city_with_type"`
				CityType             string `json:"city_type"`
				CityTypeFull         string `json:"city_type_full"`
				City                 string `json:"city"`
				CityArea             string `json:"city_area"`
				CityDistrictFiasId   any    `json:"city_district_fias_id"`
				CityDistrictKladrId  any    `json:"city_district_kladr_id"`
				CityDistrictWithType string `json:"city_district_with_type"`
				CityDistrictType     string `json:"city_district_type"`
				CityDistrictTypeFull string `json:"city_district_type_full"`
				CityDistrict         string `json:"city_district"`
				SettlementFiasId     any    `json:"settlement_fias_id"`
				SettlementKladrId    any    `json:"settlement_kladr_id"`
				SettlementWithType   any    `json:"settlement_with_type"`
				SettlementType       any    `json:"settlement_type"`
				SettlementTypeFull   any    `json:"settlement_type_full"`
				Settlement           any    `json:"settlement"`
				StreetFiasId         string `json:"street_fias_id"`
				StreetKladrId        string `json:"street_kladr_id"`
				StreetWithType       string `json:"street_with_type"`
				StreetType           string `json:"street_type"`
				StreetTypeFull       string `json:"street_type_full"`
				Street               string `json:"street"`
				SteadFiasId          any    `json:"stead_fias_id"`
				SteadCadnum          any    `json:"stead_cadnum"`
				SteadType            any    `json:"stead_type"`
				SteadTypeFull        any    `json:"stead_type_full"`
				Stead                any    `json:"stead"`
				HouseFiasId          string `json:"house_fias_id"`
				HouseKladrId         string `json:"house_kladr_id"`
				HouseCadnum          string `json:"house_cadnum"`
				HouseFlatCount       any    `json:"house_flat_count"`
				HouseType            string `json:"house_type"`
				HouseTypeFull        string `json:"house_type_full"`
				House                string `json:"house"`
				BlockType            string `json:"block_type"`
				BlockTypeFull        string `json:"block_type_full"`
				Block                string `json:"block"`
				Entrance             any    `json:"entrance"`
				Floor                any    `json:"floor"`
				FlatFiasId           any    `json:"flat_fias_id"`
				FlatCadnum           any    `json:"flat_cadnum"`
				FlatType             string `json:"flat_type"`
				FlatTypeFull         string `json:"flat_type_full"`
				Flat                 string `json:"flat"`
				FlatArea             string `json:"flat_area"`
				SquareMeterPrice     string `json:"square_meter_price"`
				FlatPrice            any    `json:"flat_price"`
				RoomFiasId           any    `json:"room_fias_id"`
				RoomCadnum           any    `json:"room_cadnum"`
				RoomType             any    `json:"room_type"`
				RoomTypeFull         any    `json:"room_type_full"`
				Room                 any    `json:"room"`
				PostalBox            any    `json:"postal_box"`
				FiasId               string `json:"fias_id"`
				FiasCode             string `json:"fias_code"`
				FiasLevel            string `json:"fias_level"`
				FiasActualityState   string `json:"fias_actuality_state"`
				KladrId              string `json:"kladr_id"`
				GeonameId            string `json:"geoname_id"`
				CapitalMarker        string `json:"capital_marker"`
				Okato                string `json:"okato"`
				Oktmo                string `json:"oktmo"`
				TaxOffice            string `json:"tax_office"`
				TaxOfficeLegal       string `json:"tax_office_legal"`
				Timezone             string `json:"timezone"`
				GeoLat               string `json:"geo_lat"`
				GeoLon               string `json:"geo_lon"`
				BeltwayHit           string `json:"beltway_hit"`
				BeltwayDistance      any    `json:"beltway_distance"`
				Metro                []struct {
					Name     string  `json:"name"`
					Line     string  `json:"line"`
					Distance float64 `json:"distance"`
				} `json:"metro"`
				Divisions     any    `json:"divisions"`
				QcGeo         string `json:"qc_geo"`
				QcComplete    any    `json:"qc_complete"`
				QcHouse       any    `json:"qc_house"`
				HistoryValues any    `json:"history_values"`
				UnparsedParts any    `json:"unparsed_parts"`
				Source        string `json:"source"`
				Qc            string `json:"qc"`
			} `json:"data"`
		} `json:"address"`
		Phones []struct {
			Value             string `json:"value"`
			UnrestrictedValue string `json:"unrestricted_value"`
			Data              struct {
				Contact     any    `json:"contact"`
				Source      string `json:"source"`
				Qc          any    `json:"qc"`
				Type        string `json:"type"`
				Number      string `json:"number"`
				Extension   any    `json:"extension"`
				Provider    string `json:"provider"`
				Country     any    `json:"country"`
				Region      string `json:"region"`
				City        any    `json:"city"`
				Timezone    string `json:"timezone"`
				CountryCode string `json:"country_code"`
				CityCode    string `json:"city_code"`
				QcConflict  any    `json:"qc_conflict"`
			} `json:"data"`
		} `json:"phones"`
		Emails []struct {
			Value             string `json:"value"`
			UnrestrictedValue string `json:"unrestricted_value"`
			Data              struct {
				Local  string `json:"local"`
				Domain string `json:"domain"`
				Type   any    `json:"type"`
				Source string `json:"source"`
				Qc     any    `json:"qc"`
			} `json:"data"`
		} `json:"emails"`
		OgrnDate      int64  `json:"ogrn_date"`
		OkvedType     string `json:"okved_type"`
		EmployeeCount int    `json:"employee_count"`
	} `json:"data"`
}
type DadataResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
}

type DadataRequest struct {
	Query string `json:"query"`
}
