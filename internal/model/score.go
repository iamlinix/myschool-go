package model

type Score struct {
	ID              int64  `gorm:"column:id;primaryKey;autoIncrement"`
	SchoolName      string `gorm:"column:school_name;type:varchar(128);uniqueIndex:score_school_suburb_state_year_grade"`
	SchoolSector    string `gorm:"column:school_sector;type:varchar(32)"`
	SchoolType      string `gorm:"column:school_type;type:varchar(16)"`
	SchoolLocation  string `gorm:"column:school_location;type:varchar(32)"`
	Suburb          string `gorm:"column:suburb;type:varchar(32);uniqueIndex:score_school_suburb_state_year_grade"`
	State           string `gorm:"column:state;type:char(3);uniqueIndex:score_school_suburb_state_year_grade"`
	Grade           string `gorm:"column:grade;type:varchar(8);uniqueIndex:score_school_suburb_state_year_grade"`
	YearRange       string `gorm:"column:year_range;type:varchar(16)"`
	Year            int    `gorm:"column:year;type:int;uniqueIndex:score_school_suburb_state_year_grade"`
	Reading         int    `gorm:"column:reading;type:int"`
	ReadingLow      int    `gorm:"column:reading_low;type:int"`
	ReadingHigh     int    `gorm:"column:reading_high;type:int"`
	ReadingSimAvg   int    `gorm:"column:reading_sim_avg;type:int"`
	ReadingSimLow   int    `gorm:"column:reading_sim_low;type:int"`
	ReadingSimHigh  int    `gorm:"column:reading_sim_high;type:int"`
	ReadingAllAvg   int    `gorm:"column:reading_all_avg;type:int"`
	Writing         int    `gorm:"column:writing;type:int"`
	WritingLow      int    `gorm:"column:writing_low;type:int"`
	WritingHigh     int    `gorm:"column:writing_high;type:int"`
	WritingSimAvg   int    `gorm:"column:writing_sim_avg;type:int"`
	WritingSimLow   int    `gorm:"column:writing_sim_low;type:int"`
	WritingSimHigh  int    `gorm:"column:writing_sim_high;type:int"`
	WritingAllAvg   int    `gorm:"column:writing_all_avg;type:int"`
	Spelling        int    `gorm:"column:spelling;type:int"`
	SpellingLow     int    `gorm:"column:spelling_low;type:int"`
	SpellingHigh    int    `gorm:"column:spelling_high;type:int"`
	SpellingSimAvg  int    `gorm:"column:spelling_sim_avg;type:int"`
	SpellingSimLow  int    `gorm:"column:spelling_sim_low;type:int"`
	SpellingSimHigh int    `gorm:"column:spelling_sim_high;type:int"`
	SpellingAllAvg  int    `gorm:"column:spelling_all_avg;type:int"`
	Grammar         int    `gorm:"column:grammar;type:int"`
	GrammarLow      int    `gorm:"column:grammar_low;type:int"`
	GrammarHigh     int    `gorm:"column:grammar_high;type:int"`
	GrammarSimAvg   int    `gorm:"column:grammar_sim_avg;type:int"`
	GrammarSimLow   int    `gorm:"column:grammar_sim_low;type:int"`
	GrammarSimHigh  int    `gorm:"column:grammar_sim_high;type:int"`
	GrammarAllAvg   int    `gorm:"column:grammar_all_avg;type:int"`
	Numeracy        int    `gorm:"column:numeracy;type:int"`
	NumeracyLow     int    `gorm:"column:numeracy_low;type:int"`
	NumeracyHigh    int    `gorm:"column:numeracy_high;type:int"`
	NumeracySimAvg  int    `gorm:"column:numeracy_sim_avg;type:int"`
	NumeracySimLow  int    `gorm:"column:numeracy_sim_low;type:int"`
	NumeracySimHigh int    `gorm:"column:numeracy_sim_high;type:int"`
	NumeracyAllAvg  int    `gorm:"column:numeracy_all_avg;type:int"`
	Total           int    `gorm:"column:total;type:int;index:score_total"`
}

func (s *Score) TableName() string {
	return "score_go"
}
