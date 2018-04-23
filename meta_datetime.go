package admin

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
)

// DatetimeConfig meta configuration used for datetime
type DatetimeConfig struct {
	MinTime  *time.Time
	MaxTime  *time.Time
	ShowTime bool
}

// ConfigureQorMeta configure datetime meta
func (datetimeConfig *DatetimeConfig) ConfigureQorMeta(metaor resource.Metaor) {
	if meta, ok := metaor.(*Meta); ok {
		timeFormat := "2006-01-02"
		if meta.Type == "datetime" {
			datetimeConfig.ShowTime = true
		}

		if datetimeConfig.ShowTime {
			timeFormat = "2006-01-02 15:04"
		}

		if meta.FormattedValuer == nil {
			meta.SetFormattedValuer(func(value interface{}, context *qor.Context) interface{} {
				switch date := meta.GetValuer()(value, context).(type) {
				case *time.Time:
					if date == nil {
						return ""
					}
					if date.IsZero() {
						return ""
					}
					return utils.FormatTime(*date, timeFormat, context)
				case time.Time:
					if date.IsZero() {
						return ""
					}
					return utils.FormatTime(date, timeFormat, context)
				default:
					return date
				}
			})
		}
	}
}

// ConfigureQORAdminFilter configure admin filter for datetime
func (datetimeConfig *DatetimeConfig) ConfigureQORAdminFilter(filter *Filter) {
	if filter.Handler == nil {
		if dbName := filter.Resource.GetMeta(filter.Name).DBName(); dbName != "" {
			filter.Handler = func(tx *gorm.DB, filterArgument *FilterArgument) *gorm.DB {
				if metaValue := filterArgument.Value.Get("Start"); metaValue != nil {
					if start := utils.ToString(metaValue.Value); start != "" {
						tx = tx.Where(fmt.Sprintf("%v > ?", dbName), start)
					}
				}

				if metaValue := filterArgument.Value.Get("End"); metaValue != nil {
					if end := utils.ToString(metaValue.Value); end != "" {
						tx = tx.Where(fmt.Sprintf("%v < ?", dbName), end)
					}
				}

				return tx
			}
		}
	}
}
