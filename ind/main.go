package main

import (
	"fmt"
	"image/color"
	"os/exec"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var com COM
var gTabIndex int

var colorRED = color.NRGBA{R: 214, G: 55, B: 55, A: 255}
var colorGREEN = color.NRGBA{R: 90, G: 210, B: 20, A: 255}
var colorBLUE = color.NRGBA{R: 80, G: 110, B: 210, A: 255}
var colorWHITE = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
var colorCream = color.NRGBA{R: 255, G: 0xFD, B: 0xD0, A: 0xFF}
var colorGray = color.NRGBA{R: 0x7C, G: 0x7C, B: 0x7C, A: 0xFF}

func main() {
	fmt.Println("Start")

	a := app.New()
	w := a.NewWindow("Программа тестирования БУ-3П")
	w.Resize(fyne.NewSize(800, 555))
	// w.SetFixedSize(true)
	w.CenterOnScreen()
	r, _ := LoadResourceFromPath("./icon.png")
	w.SetIcon(r)

	errcom := com.Open() // todo переинициализацию?
	if errcom == nil {
		defer com.Close()
	}

	com.IndsOff()

	menu := fyne.NewMainMenu(
		fyne.NewMenu("Файл"),
		// fyne.NewMenuItem("Выход (Alt+F4)", func() { a.Quit() }),
		// a quit item will be appended to our first menu
		fyne.NewMenu("Опции",
			// fyne.NewMenuItem("Параметры", nil),
			fyne.NewMenuItem("Тема", func() { changeTheme(a) }),
			// fyne.NewMenuItem("Paste", func() { fmt.Println("Menu Paste") }),
		),
		fyne.NewMenu("Справка",
			fyne.NewMenuItem("Справка", func() { aboutHelp() }),
			fyne.NewMenuItem("О программе", func() { abautProgramm() }),
		),
	)

	w.SetMainMenu(menu)
	w.SetMaster()

	tabs := container.NewAppTabs(
		container.NewTabItem("Осн. индикатор", checkMainInd()),
		container.NewTabItem("Доп. индикатор", checkAddInd()),
		container.NewTabItem("Блок реле ", checkRelayBlock()),
	)
	go func() {
		for {
			gTabIndex = tabs.CurrentTabIndex()
			// fmt.Println("TAB: ", gTabIndex)

			// version, err := com.Cmd("ver")
			// if err != nil {
			// 	com.err = err
			// } else if strings.Contains(version, "Version") {
			// 	com.err = nil
			// } else {
			// 	com.err = errors.New("нет ответа на запрос версии")
			// }

			time.Sleep(1000 * time.Millisecond)
			// runtime.Gosched()
		}
	}()

	tabs.SetTabLocation(container.TabLocationTop)
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
	err := exec.Command("cmd", "/C", ".\\help\\index.htm").Run()
	if err != nil {
		fmt.Println("Ошибка открытия файла справки")
	}
}

func abautProgramm() { // можно просто Info
	w := fyne.CurrentApp().NewWindow("О программе") // CurrentApp!
	w.Resize(fyne.NewSize(400, 200))
	w.CenterOnScreen()

	l0 := canvas.NewText("Программа проверки индикаторов", color.Black)
	l0.TextSize = 16
	l0.Move(fyne.NewPos(20, 20))
	l1 := canvas.NewText("Версия 1", color.Black)
	l0.TextSize = 14
	l2 := canvas.NewText("© 2022 ПАО «Электромеханика»", color.Black)
	text := container.NewVBox(l0, l1, l2)
	w.SetContent(fyne.NewContainerWithLayout(layout.NewCenterLayout(), text))
	w.Show() // ShowAndRun -- panic
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

	btnStart := widget.NewButton("Старт", func() {
		autoCheck = !autoCheck
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
	errorLabel.Hide()

	// отображение ошибок
	go func() {
		for {
			if com.err != nil {
				errorLabel.SetText(fmt.Sprintf("%s: %s", com.portName, com.err.Error()))
				errorLabel.Show()
			}
			/*if _, err := com.Cmd("ver"); err != nil {
				errorLabel.SetText(fmt.Sprintf("%s: %s", com.portName, err))
				errorLabel.Show()
			}*/
			if com.err == nil {
				errorLabel.Hide()
			}
			errorLabel.Refresh()
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	// проверка нажатия
	go func() {
		for {
			if gTabIndex == 0 {
				// fmt.Println("tab 1: process")
				ind1.CheckPressed()
				ind2.CheckPressed()
				ind3.CheckPressed()
			}
			time.Sleep(100 * time.Millisecond)
			runtime.Gosched()
		}
	}()

	// автоматическая проверка
	go func() {
		for {
			if (gTabIndex == 0) && autoCheck {
				// fmt.Println("tab 1: auto check START")
				btnStart.SetText("Стоп")
				for (gTabIndex == 0) && autoCheck {
					// fmt.Println("tab 1: auto check")
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
				btnStart.SetText("Старт")
			}
			time.Sleep(100 * time.Millisecond)
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
	var autoIndTest bool  // автоматическа проверка индикаторов
	var startBtnTest bool // проверка кнопок

	var btnIndStart, btnBtnStart *widget.Button

	// --------------- Индикаторы ---------------------
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

	btnIndStart = widget.NewButton("Старт", func() {
		autoIndTest = !autoIndTest
		if autoIndTest {
			btnIndStart.SetText("Стоп")
		} else {
			btnIndStart.SetText("Старт")
		}

		startBtnTest = false // не получается делать запросы из двух потоков, зависает COM todo
		btnBtnStart.SetText("Старт")

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
	errorLabel.Hide()

	// отображение ошибок
	go func() {
		for {
			if com.err == nil {
				errorLabel.Hide()
			} else {
				errorLabel.SetText(fmt.Sprintf("%s: %s", com.portName, com.err.Error()))
				errorLabel.Show()
			}
			// if com.err != nil {
			// 	errorLabel.SetText(fmt.Sprintf("%s: %s", com.portName, com.err.Error()))
			// 	errorLabel.Show()
			// } else if _, err := com.Cmd("ver"); err != nil {
			// 	errorLabel.SetText(fmt.Sprintf("%s: %s", com.portName, err))
			// 	errorLabel.Show()
			// } else if com.err == nil {
			// 	errorLabel.Hide()
			// }
			errorLabel.Refresh()
			time.Sleep(time.Second)
			// runtime.Gosched()
		}
	}()

	// автоматическая проверка
	go func() {
		for {
			if (gTabIndex == 1) && autoIndTest {
				// btnIndStart.SetText("Стоп")
				for (gTabIndex == 1) && autoIndTest {
					// fmt.Println("tab 2: auto check")
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
			time.Sleep(100 * time.Millisecond) //300
			runtime.Gosched()
		}
	}()

	indsBox := container.NewWithoutLayout(label, inds, selectbox, btnIndStart, btnIndReset)

	// ---------------------- Кнопки ---------------------

	btnBtnStart = widget.NewButton("Старт", func() {
		startBtnTest = !startBtnTest
		if startBtnTest {
			btnBtnStart.SetText("Стоп")
		} else {
			btnBtnStart.SetText("Старт")
		}

		autoIndTest = false
		btnIndStart.SetText("Старт")
	})

	var btnLight, btnP, btnT, btnContr, btnH, btnMin, btnBright BTN
	buttonsBox := container.NewGridWithColumns(
		8,
		btnBtnStart,
		btnLight.Draw(0x01, "Подсв"),
		btnP.Draw(0x02, "П"),
		btnT.Draw(0x04, "Т"),
		btnContr.Draw(0x08, "Контр"),
		btnH.Draw(0x10, "Ч"),
		btnMin.Draw(0x20, "Мин"),
		btnBright.Draw(0x40, "Ярк"),
	)

	// проверка нажатых сегментов
	go func() {
		for {
			if gTabIndex == 1 {
				// fmt.Println("tab 2: process")

				ind1.CheckPressed()
				ind2.CheckPressed()
				ind3.CheckPressed()
				ind4.CheckPressed()
				//-------------------
				// fmt.Println("tab 2: buttons")
				number, _ := com.CheckButton()
				btnLight.CheckPressed(number)
				btnP.CheckPressed(number)
				btnT.CheckPressed(number)
				btnContr.CheckPressed(number)
				btnH.CheckPressed(number)
				btnMin.CheckPressed(number)
				btnBright.CheckPressed(number)
				//-------------------
			}

			time.Sleep(100 * time.Millisecond) //200
			runtime.Gosched()

			// todo нажатие кнопок в этот поток
		}
	}()

	// проверка нажата ли кнопка на плате
	/*go func() {
		for {
			if gTabIndex == 1 && startBtnTest {
				var number int64
				// fmt.Println("tab 2: buttons")
				number, _ = com.CheckButton()
				btnLight.CheckPressed(number)
				btnP.CheckPressed(number)
				btnT.CheckPressed(number)
				btnContr.CheckPressed(number)
				btnH.CheckPressed(number)
				btnMin.CheckPressed(number)
				btnBright.CheckPressed(number)
			}
			time.Sleep(300 * time.Millisecond)
			runtime.Gosched()
		}
	}()*/

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

	// label := widget.NewLabel("") /// """"""""""""""""""""""""""
	label0 := widget.NewLabel("0")
	label1 := widget.NewLabel("1")
	label2 := widget.NewLabel("2")
	label3 := widget.NewLabel("3")
	label4 := widget.NewLabel("4")

	relayBox := container.NewGridWithColumns(
		5,
		label4, label3, label2, label1, label0,
		relay4.Draw(0x10, "4"), relay3.Draw(0x08, "3"), relay2.Draw(0x04, "2"), relay1.Draw(0x02, "1"), relay0.Draw(0x01, "0"),
	)

	btnStart := widget.NewButton("Старт", func() {
		autoCheck = !autoCheck
	})
	btnStart.Resize(fyne.NewSize(100, 40))
	btnStart.Move(fyne.NewPos(160, 330))

	errorLabel := widget.NewLabel(fmt.Sprintf("%s: Нет ошибок соединения", com.portName))
	errorLabel.Move(fyne.NewPos(420, 330))
	errorLabel.Hide()

	// box := container.NewWithoutLayout(basicLabel, selectbox, btnStart, errorLabel)

	// проверка нажатых сегментов
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
			time.Sleep(100 * time.Millisecond) //200
			runtime.Gosched()
		}
	}()

	// автоматическая проверка
	go func() {
		for {
			if (gTabIndex == 2) && autoCheck {
				fmt.Println("tab 3: auto check START")
				btnStart.SetText("Стоп")
				for _, button := range relaySlice {
					button.pressed = false
				}
				for (gTabIndex == 2) && autoCheck {
					fmt.Println("tab 3: auto check")

					for _, button := range relaySlice {
						button.pressed = true
						time.Sleep(timeout)

						for _, button := range relaySlice {
							button.pressed = false
						}
					}
				}
				btnStart.SetText("Старт")
			}
			time.Sleep(100 * time.Millisecond)
			runtime.Gosched()
		}
	}()

	box := container.NewWithoutLayout(basicLabel, selectbox, btnStart, errorLabel)
	return container.NewBorder(box, relayBox, nil, nil)

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
func (ind *IND) LightSegments(com COM) (ok bool, err error) {

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
				if ok, err := ind.LightSegments(com); err == nil {
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
				ind.LightSegments(com)
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
