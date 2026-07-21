package crmapi

// ChangelogItem — один пункт чендж-лога. Kind влияет только на визуал
// (цветной бейдж): "fixed" | "added" | "changed" | "other". CRM схлопывает
// неизвестные значения в "other", так что читателю доверять ему безопасно.
type ChangelogItem struct {
	Text string `json:"text"`
	Kind string `json:"kind"`
}

// ChangelogVersion — версия воркера и её пункты.
type ChangelogVersion struct {
	Version string `json:"version"`
	// ReleasedAt — наивная MSK-дата релиза как её хранит CRM; nil, если не
	// указана. Держим строкой, а не time.Time: поле чисто отображаемое.
	ReleasedAt *string         `json:"released_at,omitempty"`
	Items      []ChangelogItem `json:"items"`
}

// ChangelogResult — чендж-логи, которых НЕТ у клиента с переданной версией.
//
// «Актуальной версии» в системе не хранится: её роль играет максимальная метка
// в таблице чендж-логов CRM. Поэтому Latest приходит оттуда же, что и Versions,
// и рассинхрон между ними невозможен по построению.
type ChangelogResult struct {
	// Latest — актуальная версия; nil, если чендж-логов нет вообще.
	Latest *string `json:"latest,omitempty"`
	// Known — удалось ли разобрать версию клиента. false, когда воркер выключен
	// и версию не удалось получить: в этом случае утверждать «последняя версия»
	// нельзя, и Versions пуст не потому, что обновлений нет.
	Known bool `json:"known"`
	// UpToDate — клиент на актуальной версии (Known и ничего новее нет).
	UpToDate bool `json:"up_to_date"`
	// Versions — «не загруженные обновления» по возрастанию версии.
	Versions []ChangelogVersion `json:"versions"`
}
