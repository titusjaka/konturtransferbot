package konturtransferbot

import (
	"testing"
	"time"

	"github.com/ghodss/yaml"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSchedule(t *testing.T) {
	Convey("Given correct YAML schedule", t, func() {
		sYaml := []byte(`WorkDayRouteToOffice:
  - "07:30"
  - "08:00"
  - "20:00"
  - "20:30"
HolidayRouteToOffice:
  - "10:30"
WorkDayRouteFromOffice:
  - "08:20"
  - "08:50"
  - "20:20"
  - "20:50"
HolidayRouteFromOffice:
  - "18:00"`)

		Convey("It should parse into a schedule structure", func() {
			s := Schedule{}
			err := yaml.Unmarshal(sYaml, &s)
			So(err, ShouldBeNil)

			Convey("Its entries should be correct and in the same order", func() {
				So(s.WorkDayRouteToOffice[0].Format("15:04"), ShouldEqual, "07:30")
				So(s.WorkDayRouteToOffice[1].Format("15:04"), ShouldEqual, "08:00")
				So(s.WorkDayRouteToOffice[2].Format("15:04"), ShouldEqual, "20:00")
				So(s.WorkDayRouteToOffice[3].Format("15:04"), ShouldEqual, "20:30")

				So(s.HolidayRouteToOffice[0].Format("15:04"), ShouldEqual, "10:30")

				So(s.WorkDayRouteFromOffice[0].Format("15:04"), ShouldEqual, "08:20")
				So(s.WorkDayRouteFromOffice[1].Format("15:04"), ShouldEqual, "08:50")
				So(s.WorkDayRouteFromOffice[2].Format("15:04"), ShouldEqual, "20:20")
				So(s.WorkDayRouteFromOffice[3].Format("15:04"), ShouldEqual, "20:50")

				So(s.HolidayRouteFromOffice[0].Format("15:04"), ShouldEqual, "18:00")
			})

			Convey("It should correctly identify Friday as a workday", func() {
				now, _ := time.Parse("02.01.2006 15:04", "12.08.2016 07:00")
				So(s.findCorrectRoute(now, true).String(), ShouldEqual, s.WorkDayRouteToOffice.String())
				So(s.findCorrectRoute(now, false).String(), ShouldEqual, s.WorkDayRouteFromOffice.String())
			})

			Convey("It should correctly identify Sunday as a holiday", func() {
				now, _ := time.Parse("02.01.2006 15:04", "14.08.2016 07:00")
				So(s.findCorrectRoute(now, true).String(), ShouldEqual, s.HolidayRouteToOffice.String())
				So(s.findCorrectRoute(now, false).String(), ShouldEqual, s.HolidayRouteFromOffice.String())
			})

			Convey("It should recommend two best trips from office when possible", func() {
				now, _ := time.Parse("02.01.2006 15:04", "12.08.2016 07:00")
				So(s.GetBestTripFromOfficeText(now), ShouldEqual, "Ближайший дежурный рейс от офиса будет в 08:20. Следующий - в 08:50.")
			})

			Convey("It should recommend one best trip from office when possible, and notify that it is the last", func() {
				now, _ := time.Parse("02.01.2006 15:04", "12.08.2016 20:25")
				So(s.GetBestTripFromOfficeText(now), ShouldEqual, "Ближайший дежурный рейс от офиса будет в 20:50. Это последний на сегодня рейс, дальше - только на такси. "+monetizationMessage)
			})

			Convey("It should recommend to take a cab from office when no more trips are available", func() {
				now, _ := time.Parse("02.01.2006 15:04", "12.08.2016 23:00")
				So(s.GetBestTripFromOfficeText(now), ShouldEqual, "В ближайшие несколько часов уехать домой на трансфере не получится :( Придется остаться в офисе или ехать на такси. "+monetizationMessage)
			})

			Convey("It should recommend to take a cab when best available trip from office is very far away", func() {
				now, _ := time.Parse("02.01.2006 15:04", "13.08.2016 01:00")
				So(s.GetBestTripFromOfficeText(now), ShouldEqual, "В ближайшие несколько часов уехать домой на трансфере не получится :( Придется остаться в офисе или ехать на такси. "+monetizationMessage)
			})

			Convey("It should recommend two best trips to office when possible", func() {
				now, _ := time.Parse("02.01.2006 15:04", "12.08.2016 07:00")
				So(s.GetBestTripToOfficeText(now), ShouldEqual, "Ближайший дежурный рейс от Геологической будет в 07:30. Следующий - в 08:00.")
			})

			Convey("It should recommend one best trip to office when possible, and notify that it is the last", func() {
				now, _ := time.Parse("02.01.2006 15:04", "12.08.2016 20:25")
				So(s.GetBestTripToOfficeText(now), ShouldEqual, "Ближайший дежурный рейс от Геологической будет в 20:30. Это последний на сегодня рейс.")
			})

			Convey("It should recommend to get some sleep when no more trips to office are available, and recommend morning trips", func() {
				now, _ := time.Parse("02.01.2006 15:04", "12.08.2016 23:00")
				So(s.GetBestTripToOfficeText(now), ShouldEqual, "В ближайшие несколько часов уехать на работу на трансфере не получится. Лучше лечь поспать и поехать с утра. Первые рейсы от Геологической: 10:30.")
			})

			Convey("It should recommend to get some sleep when best available trip to office is very far away", func() {
				now, _ := time.Parse("02.01.2006 15:04", "10.08.2016 01:00")
				So(s.GetBestTripToOfficeText(now), ShouldEqual, "В ближайшие несколько часов уехать на работу на трансфере не получится. Лучше лечь поспать и поехать с утра. Первые рейсы от Геологической: 07:30, 08:00, 20:00.")
			})

			Convey("It should correctly return whole schedule to office", func() {
				texts := s.GetFullToOfficeTexts()
				So(texts[0], ShouldEqual, "Дежурные рейсы от Геологической в будни:\n07:30\n08:00\n20:00\n20:30\n")
				So(texts[1], ShouldEqual, "Дежурные рейсы от Геологической в выходные:\n10:30\n")
			})

			Convey("It should correctly return whole schedule from office", func() {
				texts := s.GetFullFromOfficeTexts()
				So(texts[0], ShouldEqual, "Дежурные рейсы от офиса в будни:\n08:20\n08:50\n20:20\n20:50\n")
				So(texts[1], ShouldEqual, "Дежурные рейсы от офиса в выходные:\n18:00\n")
			})
		})
	})

	Convey("Given totally invalid YAML schedule", t, func() {
		sYaml := []byte(`	1123123`)
		Convey("It should not parse", func() {
			s := Schedule{}
			err := yaml.Unmarshal(sYaml, &s)
			So(err, ShouldNotBeNil)
		})
	})
}
