package db

var primaryKeys = []string{"Date", "RowNumber"}

type Data struct {
	Date      string
	RowNumber int
	Search    string
	IsMerged  bool

	Payment             string `fieldname:"Оплата"`
	PVZ                 string `fieldname:"Код ПВЗ"`
	Email               string `fieldname:"e-mail"`
	Inscription         string `fieldname:"Надпись"`
	Details             string `fieldname:"Характеристики"`
	Texture             string `fieldname:"Фактура"`
	Pendant             string `fieldname:"Подвеска"`
	Ring                string `fieldname:"Кольцо"`
	ForNotes            string `fieldname:"Для заметок"`
	Socials             string `fieldname:"Соцсеть"`
	FullName            string `fieldname:"ФИО"`
	InscriptionBracelet string `fieldname:"Браслет надпись"`
	Description         string `fieldname:"Описание"`
	PostCode            string `fieldname:"Индекс"`
	Link                string `fieldname:"Ссылка"`
	TimeTo              string `fieldname:"...время до"`
	EdgeLower           string `fieldname:"Нижний торец"`
	DeliveryCost        string `fieldname:"Цена доставки"`
	Phone               string `fieldname:"Телефон"`
	Earrings            string `fieldname:"Серьги"`
	City                string `fieldname:"Город"`
	TimeFrom            string `fieldname:"Время с..."`
	DeliveryType        string `fieldname:"Способ доставки"`
	Notes               string `fieldname:"Заметки"`
	BoxberryNumber      string `fieldname:"Номер заказа (Boxberry)"`
	EdgeUpper           string `fieldname:"Верхний торец"`
	Type                string `fieldname:"Тип"`
	Extras              string `fieldname:"Дополнительно"`
	DeliveryAddress     string `fieldname:"Адрес доставки"`
	ForConfirmation     string `fieldname:"Для подтверждения"`
	Symbol              string `fieldname:"Символ"`
	Subtype             string `fieldname:"Вид"`
	Sum                 int    `fieldname:"Сумма"`
	PickupNumber        string `fieldname:"Номер самовывоза"`
}
