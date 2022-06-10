package main

import (
	"fmt"
	"image/color"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	storage "fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var com COM
var gVersion, gYear string
var gTabIndex int

var colorRED = color.NRGBA{R: 214, G: 55, B: 55, A: 255}
var colorGREEN = color.NRGBA{R: 90, G: 210, B: 20, A: 255}
var colorBLUE = color.NRGBA{R: 80, G: 110, B: 210, A: 255}
var colorWHITE = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
var colorCream = color.NRGBA{R: 255, G: 0xFD, B: 0xD0, A: 0xFF}
var colorGray = color.NRGBA{R: 0x7C, G: 0x7C, B: 0x7C, A: 0xFF}

func main() {
	gVersion, gYear = "1.0.0", "2022 г." // todo править при изменениях

	a := app.New()
	w := a.NewWindow("Программа проверки индикаторов")
	w.Resize(fyne.NewSize(800, 580))
	// w.SetFixedSize(true) // перестает работать заплатка для меню от quit
	w.CenterOnScreen()
	w.SetMaster()

	com.Open()
	go func() {
		for {
			if nil != com.err {
				com.Close()
				com.Open()
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer com.Close()

	com.IndsOff()

	menu := fyne.NewMainMenu(
		fyne.NewMenu("Файл",
			// a quit item will be appended to our first menu
			fyne.NewMenuItem("Тема", func() { changeTheme(a) }),
			// fyne.NewMenuItem("Выход", func() { a.Quit() }),
		),

		fyne.NewMenu("Справка",
			fyne.NewMenuItem("Посмотреть справку", func() { aboutHelp() }),
			// fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("О программе", func() { abautProgramm() }),
		),
	)
	w.SetMainMenu(menu)

	go func() { // простите
		time.Sleep(1 * time.Second)
		for _, item := range menu.Items[0].Items {
			if item.Label == "Quit" {
				item.Label = "Выход"
			}
		}
	}()

	tabs := container.NewAppTabs(
		container.NewTabItem("Осн. индикатор", checkMainInd()),
		container.NewTabItem("Доп. индикатор", checkAddInd()),
		container.NewTabItem("Блок реле ", checkRelayBlock()),
		container.NewTabItem("Инфо ", printfInfo()),
	)
	// tabs.SetTabLocation(container.TabLocationTop)
	tabs.SetTabLocation(container.TabLocationBottom)

	go func() {
		for {
			gTabIndex = tabs.CurrentTabIndex()
			time.Sleep(1000 * time.Millisecond)
			// runtime.Gosched()
		}
	}()

	w.SetContent(tabs)
	w.ShowAndRun()
}

var currentTheme bool // светлая тема false

func changeTheme(a fyne.App) {
	currentTheme = !currentTheme

	if currentTheme {
		a.Settings().SetTheme(theme.DarkTheme())
	} else {
		a.Settings().SetTheme(theme.LightTheme())
	}
}

func aboutHelp() {
	err := exec.Command("cmd", "/C", ".\\help\\index.html").Run()
	if err != nil {
		fmt.Println("Ошибка открытия файла справки")
	}
}

func abautProgramm() {
	w := fyne.CurrentApp().NewWindow("О программе") // CurrentApp!
	w.Resize(fyne.NewSize(400, 150))
	w.SetFixedSize(true)
	w.CenterOnScreen()

	img := canvas.NewImageFromURI(storage.NewFileURI("ind.png"))
	img.Resize(fyne.NewSize(66, 90)) //без изменений
	img.Move(fyne.NewPos(10, 10))

	l0 := widget.NewLabel("Программа проверки индикаторов")
	l0.Move(fyne.NewPos(80, 10))
	l1 := widget.NewLabel(fmt.Sprintf("Версия %s", gVersion))
	l1.Move(fyne.NewPos(80, 40))
	l2 := widget.NewLabel(fmt.Sprintf("© ПАО «Электромеханика», %s", gYear))
	l2.Move(fyne.NewPos(80, 70))

	box := container.NewWithoutLayout(img, l0, l1, l2)

	// w.SetContent(fyne.NewContainerWithLayout(layout.NewCenterLayout(), box))
	w.SetContent(box)
	w.Show() // ShowAndRun -- panic!
}

// ----------------------------------------------------------------------------- //
//						 Таб1: Основной индикатор								 //
// ----------------------------------------------------------------------------- //
// Индикатор для вывода скорости на БУ и БИ (3 индикатора)

func checkMainInd() fyne.CanvasObject {
	var autoCheck bool
	var timeout time.Duration // частота автоматической проверки

	label := canvas.NewText("Основной индикатор", color.Black)
	label.TextSize = 20
	label.Move(fyne.NewPos(20, 20))

	var ind1, ind2, ind3 IND
	inds := container.NewHBox(
		ind1.Draw(0x7E, 30, 80), // todo задать свои адреса
		ind2.Draw(0x7C, 190, 80),
		ind3.Draw(0x7A, 350, 80), // отключен на отладочной плате
	)

	times := []string{"0.5", "1", "2", "5"}
	selectbox := widget.NewSelect(times, func(s string) {
		timeout = convertStrToTimeout(s)
	})
	selectbox.SetSelected(times[1])
	selectbox.Resize(fyne.NewSize(100, 40))
	selectbox.Move(fyne.NewPos(30, 330))

	var btnStart *widget.Button
	btnStart = widget.NewButton("Старт", func() {
		autoCheck = !autoCheck
		if autoCheck {
			btnStart.SetText("Стоп")
		} else {
			btnStart.SetText("Старт")
		}
	})
	btnStart.Resize(fyne.NewSize(100, 40))
	btnStart.Move(fyne.NewPos(160, 330))

	btnReset := widget.NewButton("Сброс", func() {
		ind1.Reset()
		ind2.Reset()
		ind3.Reset()
		com.IndsOff()
	})
	btnReset.Resize(fyne.NewSize(100, 40))
	btnReset.Move(fyne.NewPos(290, 330))

	errorLabel := widget.NewLabel(fmt.Sprintf("%s: Нет ошибок соединения", com.portName))
	errorLabel.Move(fyne.NewPos(420, 330))

	// проверка нажатия (запись в COM, отрисовка)
	go func() {
		for {
			if gTabIndex == 0 {
				ind1.CheckPressed()
				ind2.CheckPressed()
				ind3.CheckPressed()
				com.Cmd("ver")
			}
			time.Sleep(200 * time.Millisecond)
			runtime.Gosched()
		}
	}()

	// автоматическая проверка
	go func() {
		for {
			if (gTabIndex == 0) && autoCheck {
				for (gTabIndex == 0) && autoCheck {
					for i := 0; autoCheck && (i <= 7); i++ {
						ind1.segments[i].pressed = true
						ind2.segments[i].pressed = true
						ind3.segments[i].pressed = true
						time.Sleep(timeout)
						ind1.Reset()
						ind2.Reset()
						ind3.Reset()
					}
				}
			}
			time.Sleep(100 * time.Millisecond)
			runtime.Gosched()
		}
	}()

	// отображение ошибок
	go func() {
		for {
			if gTabIndex == 0 {
				if nil == com.err {
					errorLabel.Hide()
				} else {
					errorLabel.SetText(fmt.Sprintf("%s", com.err.Error()))
					errorLabel.Show()
				}
				errorLabel.Refresh()
			}
			time.Sleep(time.Second)
			runtime.Gosched()
		}
	}()

	return container.NewWithoutLayout(label, inds, selectbox, btnStart, btnReset, errorLabel)
}

// ----------------------------------------------------------------------------- //
//					 Таб2: Дополнительный индикатор								 //
// ----------------------------------------------------------------------------- //
// Индикаторы для вывода дополнительной информации на БУ и БИ (4 индикатора),
// на этой же плате клавиатура (4 кнопки меню + 2 кн.подсветки)

func checkAddInd() fyne.CanvasObject {
	var autoIndTest bool      // автоматическа проверка индикаторов
	var timeout time.Duration // частота автоматической проверки

	label := canvas.NewText("Дополнительный индикатор", color.Black)
	label.TextSize = 20
	label.Move(fyne.NewPos(20, 20))

	var ind1, ind2, ind3, ind4 IND
	inds := container.NewHBox(
		ind1.Draw(0x7E, 30, 80),
		ind2.Draw(0x7C, 190, 80),
		ind3.Draw(0x7A, 350, 80), // отключен на отладочной плате
		ind4.Draw(0x78, 510, 80),
	)

	times := []string{"0.5", "1", "2", "5"}
	selectbox := widget.NewSelect(times, func(s string) {
		timeout = convertStrToTimeout(s)
	})
	selectbox.SetSelected(times[1])
	selectbox.Resize(fyne.NewSize(100, 40))
	selectbox.Move(fyne.NewPos(30, 330))

	var btnIndStart *widget.Button
	btnIndStart = widget.NewButton("Старт", func() {
		autoIndTest = !autoIndTest
		if autoIndTest {
			btnIndStart.SetText("Стоп")
		} else {
			btnIndStart.SetText("Старт")
		}
	})
	btnIndStart.Resize(fyne.NewSize(100, 40))
	btnIndStart.Move(fyne.NewPos(160, 330))

	btnIndReset := widget.NewButton("Сброс", func() {
		ind1.Reset()
		ind2.Reset()
		ind3.Reset()
		ind4.Reset()
		com.IndsOff()
	})
	btnIndReset.Resize(fyne.NewSize(100, 40))
	btnIndReset.Move(fyne.NewPos(290, 330))

	errorLabel := widget.NewLabel(fmt.Sprintf("%s: Нет ошибок соединения", com.portName))
	errorLabel.Move(fyne.NewPos(420, 330))

	indsBox := container.NewWithoutLayout(label, inds, selectbox, btnIndStart, btnIndReset, errorLabel)

	// автоматическая проверка индикаторов
	go func() {
		for {
			if (gTabIndex == 1) && autoIndTest {
				for (gTabIndex == 1) && autoIndTest {
					for i := 0; autoIndTest && (i <= 7); i++ {
						ind1.segments[i].pressed = true
						ind2.segments[i].pressed = true
						ind3.segments[i].pressed = true
						ind4.segments[i].pressed = true
						time.Sleep(timeout)
						ind1.Reset()
						ind2.Reset()
						ind3.Reset()
						ind4.Reset()
					}
				}
			}
			time.Sleep(100 * time.Millisecond)
			runtime.Gosched()
		}
	}()

	// Кнопки
	voidLabel := widget.NewLabel("")
	voidLabel.Hide()

	var btnLight, btnP, btnT, btnContr, btnH, btnMin, btnBright BTN
	buttonsBox := container.NewGridWithColumns(
		7,
		btnLight.Draw(0x01, "Подсв"),
		btnP.Draw(0x02, "П"),
		btnT.Draw(0x04, "Т"),
		btnContr.Draw(0x08, "Контр"),
		btnH.Draw(0x10, "Ч"),
		btnMin.Draw(0x20, "Мин"),
		btnBright.Draw(0x40, "Ярк"),
		container.NewWithoutLayout(voidLabel),
	)

	// проверка нажатых элементов (запись в COM, отрисовка)
	go func() {
		for {
			if gTabIndex == 1 {
				ind1.CheckPressed()
				ind2.CheckPressed()
				ind3.CheckPressed()
				ind4.CheckPressed()

				number, _ := com.CheckButton()
				btnLight.CheckPressed(number)
				btnP.CheckPressed(number)
				btnT.CheckPressed(number)
				btnContr.CheckPressed(number)
				btnH.CheckPressed(number)
				btnMin.CheckPressed(number)
				btnBright.CheckPressed(number)
			}
			time.Sleep(200 * time.Millisecond)
			runtime.Gosched()
		}
	}()

	// отображение ошибок
	go func() {
		for {
			if gTabIndex == 1 {
				if nil == com.err {
					errorLabel.Hide()
				} else {
					errorLabel.SetText(fmt.Sprintf("%s", com.err.Error()))
					errorLabel.Show()
				}
				errorLabel.Refresh()
			}
			time.Sleep(time.Second)
			runtime.Gosched()
		}
	}()

	return container.NewBorder(indsBox, buttonsBox, nil, nil)
}

func convertStrToTimeout(s string) (t time.Duration) {
	switch s {
	case "0.5":
		t = time.Second / 2
	case "1":
		t = time.Second
	case "2":
		t = 2 * time.Second
	case "5":
		t = 5 * time.Second
	}
	return
}

// ----------------------------------------------------------------------------- //
//							 Таб3:	Блок реле		 							 //
// ----------------------------------------------------------------------------- //
// Плата с релюхами (5 реле) уставок скоростей БУ

func checkRelayBlock() fyne.CanvasObject {
	var autoCheck bool
	var timeout time.Duration // частота автоматической проверки

	basicLabel := canvas.NewText("Блок реле", color.Black)
	basicLabel.TextSize = 20
	basicLabel.Move(fyne.NewPos(20, 20))

	times := []string{"0.5", "1", "2", "5"}
	selectbox := widget.NewSelect(times, func(s string) {
		timeout = convertStrToTimeout(s)
	})
	selectbox.SetSelected(times[1])
	selectbox.Resize(fyne.NewSize(100, 40))
	selectbox.Move(fyne.NewPos(30, 330))

	var relay0, relay1, relay2, relay3, relay4 BTN
	relaySlice := []*BTN{&relay0, &relay1, &relay2, &relay3, &relay4}

	relayBox := container.NewGridWithColumns(
		5,
		relay4.Draw(0x10, "4"), relay3.Draw(0x08, "3"), relay2.Draw(0x04, "2"), relay1.Draw(0x02, "1"), relay0.Draw(0x01, "0"),
	)

	btnStart := widget.NewButton("Старт", func() {
		autoCheck = !autoCheck
	})
	btnStart.Resize(fyne.NewSize(100, 40))
	btnStart.Move(fyne.NewPos(160, 330))

	errorLabel := widget.NewLabel(fmt.Sprintf("%s: Нет ошибок соединения", com.portName))
	errorLabel.Move(fyne.NewPos(300, 330))
	errorLabel.Hide()

	// проверка нажатых сегментов (запись в COM, отрисовка)
	go func() {
		var pressedRelays int // все реле установленные в единицу
		changedbit := false
		for {
			if gTabIndex == 2 {
				// fmt.Println("tab 3: process")
				for _, button := range relaySlice {
					if button.pressed {
						if (pressedRelays & button.number) != button.number {
							pressedRelays |= button.number
							changedbit = true
						}
					} else {
						if (pressedRelays & button.number) == button.number {
							pressedRelays &= ^button.number
							changedbit = true
						}
					}
					if changedbit {
						changedbit = false
						if setbits, err := com.CheckRelay(pressedRelays); err == nil {
							button.CheckPressed(setbits)
						} else {
							// ошибка передачи COM
						}
					}
				}
			}
			time.Sleep(100 * time.Millisecond)
			runtime.Gosched()
		}
	}()

	// автоматическая проверка
	go func() {
		for {
			if (gTabIndex == 2) && autoCheck {
				// fmt.Println("tab 3: auto check START")
				btnStart.SetText("Стоп")
				for _, button := range relaySlice { // если натыкали ручками убирам
					button.pressed = false
				}
				for (gTabIndex == 2) && autoCheck {
					// fmt.Println("tab 3: auto check")

					for _, button := range relaySlice {
						button.pressed = true
						time.Sleep(timeout)

						for _, button := range relaySlice {
							button.pressed = false
						}
						if !autoCheck {
							break
						}
					}
				}
				btnStart.SetText("Старт")
			}
			time.Sleep(100 * time.Millisecond)
			runtime.Gosched()
		}
	}()

	// отображение ошибок
	go func() {
		for {
			if gTabIndex == 2 {
				if nil == com.err {
					errorLabel.Hide()
				} else {
					errorLabel.SetText(fmt.Sprintf("%s", com.err.Error()))
					errorLabel.Show()
				}
				errorLabel.Refresh()
			}
			time.Sleep(time.Second)
			runtime.Gosched()
		}
	}()

	voidLabel := widget.NewLabel("")

	box := container.NewWithoutLayout(basicLabel, selectbox, btnStart, errorLabel)
	return container.NewVBox(box, voidLabel, voidLabel, voidLabel, voidLabel, relayBox) // не понятно как делать разметку, использую пустой лейбл чтобы опустить контейнер ниже
}

// ----------------------------------------------------------------------------- //
//							 Таб4:	Информация		 							 //
// ----------------------------------------------------------------------------- //

func printfInfo() fyne.CanvasObject {
	versionPBI := "номер версии не получен" //Версия программы платы интерфейсной

	title := canvas.NewText("Информация", color.Black)
	title.TextSize = 20
	title.Move(fyne.NewPos(20, 20))

	label := widget.NewLabel("")
	versionLabel := widget.NewLabel("Версия программы: " + gVersion)
	comLabel := widget.NewLabel("")
	versionPBILabel := widget.NewLabel("")

	box := container.NewVBox(label, label, versionLabel, comLabel, versionPBILabel)

	go func() {
		for {
			if gTabIndex == 3 {
				versionPBI, _ = com.Cmd("ver")
			}
			time.Sleep(500 * time.Millisecond)
			runtime.Gosched()
		}
	}()

	// отображение ошибок и версии
	go func() {
		for {
			if gTabIndex == 3 {
				if "" == versionPBI || !strings.Contains(versionPBI, "Version") {
					versionPBILabel.SetText("Ошибка получения версии программы платы интерфейсной")
				} else {
					ver := strings.Trim(versionPBI, "Version ")
					versionPBILabel.SetText("Версия программы платы интерфейсной: " + ver)
				}
				versionPBILabel.Refresh()

				if nil == com.err {
					comLabel.SetText(fmt.Sprintf("Соединение установлено: %s", com.portName))
				} else {
					comLabel.SetText(fmt.Sprintf("%s", com.err.Error()))
					// errorLabel.Show()
				}
				comLabel.Refresh()
			}
			time.Sleep(time.Second)
			runtime.Gosched()
		}
	}()

	return container.NewWithoutLayout(title, box)
}

// ----------------------------------------------------------------------------- //
//					 Отрисовка индикатора (8 сегментов)							 //
// ----------------------------------------------------------------------------- //

// IND индикатор
type IND struct {
	number      int // 78 7A 7C 7E указывает индикатор
	litSegments int // все выделенные сегменты
	segments    [8]SEG
}

// Draw отрисовка
//  x, y - смещение индикатора относительно
func (ind *IND) Draw(number int, x, y float32) *fyne.Container {
	ind.number = number

	// отрисовка  левого  верхнего сегмента
	s0 := ind.segments[0].Draw(0x40, x+30, y+0)
	s2 := ind.segments[1].Draw(0x20, x+90, y+30)
	s5 := ind.segments[2].Draw(0x10, x+90, y+120)
	s6 := ind.segments[3].Draw(0x08, x+30, y+180)
	s4 := ind.segments[4].Draw(0x04, x+0, y+120)
	s1 := ind.segments[5].Draw(0x01, x+0, y+30)
	s3 := ind.segments[6].Draw(0x02, x+30, y+90)
	s7 := ind.segments[7].Draw(0x80, x+120, y+180)

	return container.NewWithoutLayout(
		s0, s1, s2, s3, s4, s5, s6, s7,
	)
}

// Hide только для кнопки общего сброса
func (ind *IND) Hide() {

	for i := 0; i < len(ind.segments); i++ {
		ind.segments[i].Hide()
	}
}

// LightSegments зажечь выбранные(litSegments) сегменты не плате
func (ind *IND) LightSegments() (ok bool, err error) {

	cmd := "w" + fmt.Sprintf("%X=", ind.number) + fmt.Sprintf("%X", ind.litSegments) // w78=01
	ok, err = com.CheckInd(cmd)
	return
}

/*	ручная проверка
	Проверка всех 8 сегментов:
	если сегмент был нажат (pressed), проверить не был ли он уже подсвечен (не спамить в com),
		подсветить на плате, отметить в окне программы
	если сегмент был сброшен, проверить не сброшен ли он уже,
		сбросить на плате, убрать выделение в окне программы
*/

// CheckPressed проверка сегментов
func (ind *IND) CheckPressed() error {

	for i := 0; i <= 7; i++ {
		seg := ind.segments[i].number

		if ind.segments[i].pressed {
			if (ind.litSegments & seg) != seg {

				ind.litSegments |= seg
				if ok, err := ind.LightSegments(); err == nil {
					if ok {
						ind.segments[i].ShowGreen()
					} else if !ok {
						ind.segments[i].ShowRed()
					}
				}
			}
		} else {
			if (ind.litSegments & seg) == seg {
				ind.litSegments &= ^seg
				ind.LightSegments()
				ind.segments[i].Hide()
			}
		}
	}
	return nil
}

// Reset очистить все сегменты
func (ind *IND) Reset() {
	ind.Hide()
	for i := 0; i <= 7; i++ {
		ind.segments[i].pressed = false
	}
}

// ----------------------------------------------------------------------------- //
//					 Отрисовка сегмента индикатора								 //
// ----------------------------------------------------------------------------- //

/* номера сегментов:
  0x40
0x    0x
01    20
  0x02
0x    0x
04    10
  0x08  x80
*/

// SEG сегмент
type SEG struct {
	number    int
	pos       fyne.Position
	rectGreen *canvas.Rectangle
	rectRed   *canvas.Rectangle
	pressed   bool
}

func getSegmentSize(number int) (s fyne.Size) {

	switch number {
	case 0x40, 0x02, 0x08:
		s = fyne.NewSize(60, 30) // для горизонтально расположенных
	case 0x01, 0x20, 0x04, 0x10:
		s = fyne.NewSize(30, 60) // для вертикальных
	case 0x80:
		s = fyne.NewSize(30, 30) // точка
	default:
		fmt.Println("getSegmentSize(): ERROR!!!")
	}
	return
}

// Draw отрисовка
// number	- номер сегмента
// x, y 	- смещение от начала координат
func (seg *SEG) Draw(number int, x, y float32) *fyne.Container {
	seg.number = number

	size := getSegmentSize(number)
	seg.pos = fyne.NewPos(x, y)
	seg.rectGreen = canvas.NewRectangle(colorGREEN)
	seg.rectRed = canvas.NewRectangle(colorRED)
	btn := widget.NewButton("", func() {
		seg.pressed = !seg.pressed
	})
	btn.Resize((size))
	btn.Move((seg.pos))

	box := container.NewWithoutLayout(
		btn, seg.rectGreen, seg.rectRed,
	)
	return box
}

// ShowGreen отметить сегмент как нажатый
func (seg *SEG) ShowGreen() {
	size := getSegmentSize(seg.number)
	seg.rectGreen.Resize(size)
	seg.rectGreen.Move(seg.pos)
	seg.rectGreen.Show()
	seg.rectGreen.Refresh()

}

// ShowRed отметить сегмент как ошибку
func (seg *SEG) ShowRed() {
	size := getSegmentSize(seg.number)
	seg.rectRed.Resize(size)
	seg.rectRed.Move(seg.pos)
	seg.rectRed.Show()
	seg.rectRed.Refresh()
}

// Hide неподсвечивать (сегмент не нажат)
func (seg *SEG) Hide() {
	seg.rectRed.Hide()
	seg.rectGreen.Hide()
}

// ----------------------------------------------------------------------------- //
//									 Кнопки								 		 //
// ----------------------------------------------------------------------------- //

/* номера:
подсветка (лампочка)= 0x1
Дата П 				= 0x2
Время Т 			= 0x4
Путь Контр 			= 0x8
Лента Ч 			= 0x10
Вверх Мин 			= 0x20
яркость (солнышко)	= 0x40
*/

// BTN кнопка
type BTN struct {
	number      int
	button      *widget.Button
	rectPressed *canvas.Rectangle
	rectErr     *canvas.Rectangle
	pressed     bool
}

// Draw отрисовка
// number - номер сегмента
// x, y - смещение от начала координат
func (btn *BTN) Draw(number int, name string) *fyne.Container {

	btn.number = number
	btn.button = widget.NewButton(name, func() {
		btn.pressed = !btn.pressed
	})
	btn.rectPressed = canvas.NewRectangle(colorBLUE)
	btn.rectErr = canvas.NewRectangle(colorRED)
	btn.rectPressed.Hide()
	btn.rectPressed.Refresh()
	btn.rectErr.Hide()
	btn.rectErr.Refresh()

	box := container.NewPadded(
		btn.rectPressed, btn.rectErr, btn.button,
	)

	return box
}

// CheckPressed проверяет было ли нажатие на кнопку на плате
// отмечается нажатая кнопка на экране
// number -- все нажатые на текущий момент кнопки
func (btn *BTN) CheckPressed(number int64) {

	if number == -1 {
		btn.ShowErr()
	} else if (int(number) & btn.number) == btn.number {
		btn.Show()
	} else {
		btn.Hide()
	}
}

// ShowErr отметить кнопку как ошибку
func (btn *BTN) ShowErr() {
	btn.rectErr.Show()
	btn.rectErr.Refresh()
	btn.button.Refresh()
}

// Show отметить кнопку как нажатую
func (btn *BTN) Show() {
	btn.rectPressed.Show()
	btn.rectPressed.Refresh()
	btn.button.Refresh()
}

// Hide не подсвечивать кнопки
func (btn *BTN) Hide() {
	btn.rectErr.Hide()
	btn.rectErr.Refresh()

	btn.rectPressed.Hide()
	btn.rectPressed.Refresh()

	btn.button.Refresh()
}

// ----------------------------------------------------------------------------- //
