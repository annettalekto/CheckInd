package main

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var com COM

// var ColorRED = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
var ColorRED = color.NRGBA{R: 214, G: 55, B: 55, A: 255} // чуть бледнее
var ColorGREEN = color.NRGBA{R: 90, G: 210, B: 20, A: 255}
var ColorBLUE = color.NRGBA{R: 80, G: 110, B: 210, A: 255}
var ColorWHITE = color.NRGBA{R: 255, G: 255, B: 255, A: 255}

func main() {
	fmt.Println("Start")

	com.Open()
	defer com.Close()
	allIndsOff(com) // переименовать и переделать todo

	a := app.New()
	w := a.NewWindow("Программа тестирования БУ-3П")
	w.Resize(fyne.NewSize(800, 540))
	// w.SetFixedSize(true)

	menu := fyne.NewMainMenu(
		fyne.NewMenu("Файл"),
		// fyne.NewMenuItem("Выход (Alt+F4)", func() { a.Quit() }),
		// a quit item will be appended to our first menu
		fyne.NewMenu("Опции",
			fyne.NewMenuItem("Параметры", nil),
			fyne.NewMenuItem("Тема", func() { changeTheme(a) }),
			// fyne.NewMenuItem("Paste", func() { fmt.Println("Menu Paste") }),
		),
		fyne.NewMenu("Справка",
			fyne.NewMenuItem("Пункт 1", func() { fmt.Println("Что-то 1") }),
		),
	)

	w.SetMainMenu(menu)
	w.SetMaster()

	tabs := container.NewAppTabs(
		container.NewTabItem("Осн. индикатор", checkMainInd()),
		container.NewTabItem("Доп. индикатор", checkAddInd()),
		container.NewTabItem("Блок реле ", checkRelayBlock()),
	)

	tabs.SetTabLocation(container.TabLocationTop)
	w.SetContent(tabs)

	w.ShowAndRun()
}

var currentTheme bool

func changeTheme(a fyne.App) {
	if currentTheme {
		a.Settings().SetTheme(theme.DarkTheme())
		currentTheme = false
	} else {
		a.Settings().SetTheme(theme.LightTheme())
		currentTheme = true
	}
}

// ----------------------------------------------------------------------------- //
//						 Таб1: Основной индикатор								 //
// ----------------------------------------------------------------------------- //

func checkMainInd() fyne.CanvasObject {
	var autoCheck bool
	var timeout time.Duration // частота автоматической проверки

	label := canvas.NewText("Основной индикатор", color.Black)
	label.TextSize = 20
	label.Move(fyne.NewPos(20, 20))

	var ind1, ind2, ind3 IND
	inds := container.NewHBox(
		ind1.Draw(0x7E, 30, 80), //todo задать свои адреса
		ind2.Draw(0x7C, 190, 80),
		ind3.Draw(0x7A, 350, 80), // отключен на отладочной плате
	)

	times := []string{"0.5", "1", "2", "5"}
	selectbox := widget.NewSelect(times, func(s string) {
		timeout = convertStrToTimeout(s)
	})
	selectbox.SetSelected(times[1])
	selectbox.Resize(fyne.NewSize(100, 30))
	selectbox.Move(fyne.NewPos(30, 330))

	btnStart := widget.NewButton("Старт", func() {
		autoCheck = !autoCheck
	})
	btnStart.Resize(fyne.NewSize(100, 30))
	btnStart.Move(fyne.NewPos(180, 330))

	btnReset := widget.NewButton("Сброс", func() {
		ind1.Reset()
		ind2.Reset()
		ind3.Reset()
		allIndsOff(com)
	})
	btnReset.Resize(fyne.NewSize(100, 30))
	btnReset.Move(fyne.NewPos(320, 330))

	// ручная проверка
	go func() {
		for {
			ind1.CheckPressed()
			ind2.CheckPressed()
			ind3.CheckPressed()
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// автоматическая проверка
	go func() {
		for {
			if autoCheck {
				btnStart.SetText("Стоп")
				for autoCheck {
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
		}
	}()

	return container.NewWithoutLayout(label, inds, selectbox, btnStart, btnReset)
}

// ----------------------------------------------------------------------------- //
//					 Таб2: Дополнительный индикатор								 //
// ----------------------------------------------------------------------------- //

func checkAddInd() fyne.CanvasObject {

	buttonsBox := buttons()
	indsBox := indicators()

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

// отрисовка на форме индикаторов и кнопок к ним, обработка нажатия
func indicators() *fyne.Container {
	var autoCheck bool
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
	selectbox.Resize(fyne.NewSize(100, 30))
	selectbox.Move(fyne.NewPos(30, 330))

	btnStart := widget.NewButton("Старт", func() {
		autoCheck = !autoCheck
	})
	btnStart.Resize(fyne.NewSize(100, 30))
	btnStart.Move(fyne.NewPos(180, 330))

	btnReset := widget.NewButton("Сброс", func() {
		ind1.Reset()
		ind2.Reset()
		ind3.Reset()
		ind4.Reset()
		allIndsOff(com)
	})
	btnReset.Resize(fyne.NewSize(100, 30))
	btnReset.Move(fyne.NewPos(320, 330))

	// ручная проверка
	go func() {
		for {
			ind1.CheckPressed()
			ind2.CheckPressed()
			ind3.CheckPressed()
			ind4.CheckPressed()
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// автоматическая проверка
	go func() {
		for {
			if autoCheck {
				btnStart.SetText("Стоп")
				for autoCheck {
					for i := 0; autoCheck && (i <= 7); i++ {
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
				btnStart.SetText("Старт")
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	box := container.NewWithoutLayout(label, inds, selectbox, btnStart, btnReset)

	return box
}

// отрисовка на форме кнопок, обработка нажатия
func buttons() *fyne.Container {

	var btnLight, btnP, btnT, btnContr, btnH, btnMin, btnBright BTN
	grid := container.NewGridWithColumns(
		7,
		btnLight.Draw(0x01, "Подсв"),
		btnP.Draw(0x02, "П"),
		btnT.Draw(0x04, "Т"),
		btnContr.Draw(0x08, "Контр"),
		btnH.Draw(0x10, "Ч"),
		btnMin.Draw(0x20, "Мин"),
		btnBright.Draw(0x40, "Ярк"),
	)

	// проверка нажата ли кнопка на плате
	go func() {
		for {
			number, _ := ButtonOn(com)
			btnLight.CheckPressed(number)
			btnP.CheckPressed(number)
			btnT.CheckPressed(number)
			btnContr.CheckPressed(number)
			btnH.CheckPressed(number)
			btnMin.CheckPressed(number)
			btnBright.CheckPressed(number)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	return grid
}

// ----------------------------------------------------------------------------- //
//							 Таб3:	Блок реле		 							 //
// ----------------------------------------------------------------------------- //

func checkRelayBlock() fyne.CanvasObject {
	basicLabel := canvas.NewText("Блок реле", color.Black)
	basicLabel.TextSize = 20
	basicLabel.Move(fyne.NewPos(20, 20))

	btn0 := widget.NewButton("0", nil)
	btn1 := widget.NewButton("0", nil)
	btn2 := widget.NewButton("0", nil)
	btn3 := widget.NewButton("0", nil)
	btn4 := widget.NewButton("0", nil)

	label := widget.NewLabel("")
	label0 := widget.NewLabel("0")
	label1 := widget.NewLabel("1")
	label2 := widget.NewLabel("2")
	label3 := widget.NewLabel("3")
	label4 := widget.NewLabel("4")

	grid := container.NewGridWithColumns(
		5,
		label0, label1, label2, label3, label4,
		btn0, btn1, btn2, btn3, btn4,
	)

	btnStart := widget.NewButton("Старт", nil)
	btnStart.Resize(fyne.NewSize(100, 30))
	btnStart.Move(fyne.NewPos(20, 330))

	box0 := container.NewWithoutLayout(basicLabel, btnStart)

	return container.NewVBox(box0, label, grid)
}

// ----------------------------------------------------------------------------- //
//					 Отрисовка индикатора (8 сегментов)							 //
// ----------------------------------------------------------------------------- //

type IND struct {
	number      int // 78 7A 7C 7E указывает индикатор
	litSegments int // все выделенные сегменты
	segments    [8]SEG
}

// x, y - смещение индикатора относительно
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

// func (ind *IND) Show() { // нужно знать красным или зеленым

// 	for i := 0; i < len(ind.segments); i++ {
// 		ind.segments[i].Show()
// 	}
// }

// только для кнопки общего сброса
func (ind *IND) Hide() {

	for i := 0; i < len(ind.segments); i++ {
		ind.segments[i].Hide()
	}
}

// зажечь на плате сегменты, которые были выбраны (litSegments)
// (зажигать последний выбранный, но не гасить другие)
func (ind *IND) LightSegments(com COM) (ok bool) {

	cmd := "w" + fmt.Sprintf("%X=", ind.number) + fmt.Sprintf("%X", ind.litSegments) // w78=01
	ok, _ = IndOn(com, cmd)
	return
}

/*	Ручная проверка
	Проверка всех 8 сегментов:
	если сегмент был нажат (pressed), проверить не был ли он уже подсвечен (не спамить в com),
		подсветить на плате, отметить в окне программы
	если сегмент был сброшен, проверить не сброшен ли он уже,
		сбросить на плате, убрать выделение в окне программы
*/
func (ind *IND) CheckPressed() {

	for i := 0; i <= 7; i++ {
		seg := ind.segments[i].number

		if ind.segments[i].pressed {
			if (ind.litSegments & seg) != seg {

				ind.litSegments |= seg
				if ind.LightSegments(com) {
					ind.segments[i].ShowGreen()
				} else {
					ind.segments[i].ShowRed()
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
}

func (ind *IND) Reset() {
	ind.Hide()
	for i := 0; i <= 7; i++ {
		ind.segments[i].pressed = false
		ind.segments[i].pressed = false
		ind.segments[i].pressed = false
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

// Сегмент
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

// number - номер сегмента
// x, y - смещение от начала координат
func (seg *SEG) Draw(number int, x, y float32) *fyne.Container {
	green := ColorGREEN
	red := ColorRED

	seg.number = number

	size := getSegmentSize(number)
	seg.pos = fyne.NewPos(x, y)
	seg.rectGreen = canvas.NewRectangle(green)
	seg.rectRed = canvas.NewRectangle(red)
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

func (seg *SEG) ShowGreen() {
	size := getSegmentSize(seg.number)
	seg.rectGreen.Resize(size)
	seg.rectGreen.Move(seg.pos)
	seg.rectGreen.Show()
	seg.rectGreen.Refresh()

}

func (seg *SEG) ShowRed() {
	size := getSegmentSize(seg.number)
	seg.rectRed.Resize(size)
	seg.rectRed.Move(seg.pos)
	seg.rectRed.Show()
	seg.rectRed.Refresh()
}

func (seg *SEG) Hide() {
	seg.rectRed.Hide()
	seg.rectGreen.Hide()
}

/*
// Зажечь на плате, подсветить кнопку зеленым или красным
// (для автоматической проверки сегментов покругу)
func (seg *SEG) CheckSegment(com COM, indNumber int) (ok bool) {

	cmd := "w" + fmt.Sprintf("%X=", indNumber) + fmt.Sprintf("%X", seg.number) // w78=40

	ok, _ = IndOn(com, cmd)
	if ok {
		seg.ShowGreen()
	} else {
		seg.ShowRed()
	}
	return
}*/

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

// Сегмент
type BTN struct {
	number    int
	rectBlue  *canvas.Rectangle
	rectRed   *canvas.Rectangle
	rectWhite *canvas.Rectangle
	showed    bool
}

// number - номер сегмента
// x, y - смещение от начала координат
func (btn *BTN) Draw(number int, name string) *fyne.Container {
	red := ColorRED
	blue := ColorBLUE
	white := ColorWHITE

	btn.number = number
	button := widget.NewButton(name, nil)
	btn.rectBlue = canvas.NewRectangle(blue)
	btn.rectRed = canvas.NewRectangle(red)
	btn.rectWhite = canvas.NewRectangle(white) // без белого треугольника вылезат красный  при NewPadded() почему то
	// btn.rectBlue.Hide() // если скрыть сразу отрисовка тупит
	// btn.rectRed.Hide()

	box := container.NewPadded(
		btn.rectBlue, btn.rectRed, btn.rectWhite, button,
	)

	return box
}

// проверяется нажатие на кнопку на плате
// отмечается нажатая кнопка на экране
func (btn *BTN) CheckPressed(number int64) {

	if number == -1 {
		btn.ShowRed()
	}
	if (int(number) & btn.number) == btn.number { // если кнопка нажата
		if !btn.showed { // не отрисовывать лишний раз
			btn.showed = true
			btn.ShowBlue()
		}
	} else { // не нажата
		if btn.showed {
			btn.showed = false
			btn.Hide()
		}
	}
}

func (btn *BTN) ShowRed() {
	btn.rectBlue.Hide()
	btn.rectWhite.Hide()

	btn.rectRed.Show()
	btn.rectRed.Refresh()
}

func (btn *BTN) ShowBlue() {
	btn.rectWhite.Hide()
	btn.rectRed.Hide()

	btn.rectBlue.Show()
	btn.rectBlue.Refresh()
}

func (btn *BTN) Hide() {
	btn.rectRed.Hide()
	btn.rectBlue.Hide()

	btn.rectWhite.Show()
	btn.rectWhite.Refresh()
}

// ----------------------------------------------------------------------------- //
