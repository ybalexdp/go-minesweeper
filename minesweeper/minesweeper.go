package minesweeper

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	column   int
	row      int
	bombnum  int
	gameover bool
	remain   int
	num      [9]string
	points   [][]point
}

type point struct {
	data    int
	opened  bool
	flagged bool
}

const (
	cellHeight = 2
	cellWidth  = 5

	marginLeft = 3
	marginTop  = 2

	vertical   = "│"
	horizontal = "──"

	minRow  = 5
	maxRow  = 25
	minCol  = 5
	maxCol  = 25
	minBomb = 1

	flagIcon = " 🚩 "
	bombIcon = " 💣 "

	gameoverView = `

 ██████╗  █████╗ ███╗   ███╗███████╗ ██████╗ ██╗   ██╗███████╗██████╗ 
██╔════╝ ██╔══██╗████╗ ████║██╔════╝██╔═══██╗██║   ██║██╔════╝██╔══██╗
██║  ███╗███████║██╔████╔██║█████╗  ██║   ██║██║   ██║█████╗  ██████╔╝
██║   ██║██╔══██║██║╚██╔╝██║██╔══╝  ██║   ██║╚██╗ ██╔╝██╔══╝  ██╔══██╗
╚██████╔╝██║  ██║██║ ╚═╝ ██║███████╗╚██████╔╝ ╚████╔╝ ███████╗██║  ██║
 ╚═════╝ ╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝ ╚═════╝   ╚═══╝  ╚══════╝╚═╝  ╚═╝
`

	gameclearView = `

 ██████╗██╗     ███████╗ █████╗ ██████╗ ██╗██╗
██╔════╝██║     ██╔════╝██╔══██╗██╔══██╗██║██║
██║     ██║     █████╗  ███████║██████╔╝██║██║
██║     ██║     ██╔══╝  ██╔══██║██╔══██╗╚═╝╚═╝
╚██████╗███████╗███████╗██║  ██║██║  ██║██╗██╗
 ╚═════╝╚══════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝╚═╝
 `
)

func InitialModel() (model, error) {

	m := model{}
	var err error

	fmt.Print("Enter the number of columns[5-25]:")
	sc := bufio.NewScanner(os.Stdin)
	sc.Scan()

	m.column, err = strconv.Atoi(sc.Text())
	if err != nil {
		return m, fmt.Errorf("The number of columns must be in the range of %d-%d", minCol, maxCol)
	}

	if m.column > maxCol || m.column < minCol {
		return m, fmt.Errorf("The number of columns must be in the range of %d-%d", minCol, maxCol)
	}

	fmt.Print("Enter the number of rows[5-25]:")
	sc.Scan()
	m.row, err = strconv.Atoi(sc.Text())

	if err != nil {
		return m, fmt.Errorf("The number of rows must be in the range of %d-%d", minRow, maxRow)
	}

	if m.row > maxRow || m.row < minRow {
		return m, fmt.Errorf("The number of rows must be in the range of %d-%d", minRow, maxRow)
	}

	maxBomb := m.row*m.column - 1

	fmt.Printf("Enter the number of bombs[1-%d]:", maxBomb)
	sc.Scan()

	m.bombnum, err = strconv.Atoi(sc.Text())
	if err != nil {
		return m, fmt.Errorf("The number of bombs must be in the range of %d-%d", minBomb, maxBomb)
	}

	if m.bombnum > maxBomb || m.bombnum < minBomb {
		return m, fmt.Errorf("The number of bombs must be in the range of %d-%d", minBomb, maxBomb)
	}

	m.points = make([][]point, m.row)
	for i := range m.points {
		m.points[i] = make([]point, m.column)
	}

	i := 0
	for i < m.bombnum {
		rand.Seed(time.Now().UnixNano())
		x := rand.Intn(m.row)
		y := rand.Intn(m.column)

		if m.points[x][y].data == -1 {
			continue
		}

		m.points[x][y].data = -1
		for cx := x - 1; cx < m.row && cx <= x+1; cx++ {
			for cy := y - 1; cy < m.column && cy <= y+1; cy++ {
				if cx < 0 || cy < 0 || (cx == x && cy == y) || m.points[cx][cy].data == -1 {
					continue
				}
				m.points[cx][cy].data++
			}
		}
		i++
	}

	m.remain = m.row*m.column - m.bombnum

	m.setNumView()

	return m, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.MouseMsg:
		if m.gameover || m.remain == 0 {
			return m, nil
		}
		col, row := m.cell(msg.X, msg.Y)
		if col == -1 || row == -1 {
			return m, nil
		}
		if msg.Alt && msg.Type == tea.MouseLeft && !m.points[row][col].opened {
			if m.points[row][col].flagged {
				m.points[row][col].flagged = !m.points[row][col].flagged
				m.bombnum++
			} else if m.bombnum > 0 {
				m.bombnum--
				m.points[row][col].flagged = !m.points[row][col].flagged
			}
			return m, nil
		}

		if msg.Type != tea.MouseLeft {
			return m, nil
		}
		return m.choose(col, row)

	case tea.KeyMsg:

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	var s strings.Builder
	s.WriteString(m.viewHeader())
	s.WriteString(top(m.column - 1))

	for i := 0; i < m.row; i++ {
		s.WriteString(cellLeft())
		for j := 0; j < m.column; j++ {
			if m.gameover {
				m.points[i][j].opened = true
			}
			s.WriteString(m.cellMiddle(m.points[i][j].data, m.points[i][j].opened, m.points[i][j].flagged))

		}

		s.WriteString(cellRight())

		if i < m.row-1 {
			s.WriteString(middle(m.column - 1))
		} else {
			s.WriteString(bottom(m.column - 1))
		}
	}

	if m.gameover {
		s.WriteString(gameover())
	} else if m.remain == 0 {
		s.WriteString(gameclear())
	}

	return s.String()
}

func (m *model) multiOpen(row, col int) {

	for x := row - 1; x <= row+1 && x < m.row; x++ {
		for y := col - 1; y <= col+1 && y < m.column; y++ {
			if x < 0 || y < 0 || m.points[x][y].opened {
				continue
			}
			m.points[x][y].opened = true
			m.remain--
			if m.points[x][y].data == 0 {
				m.multiOpen(x, y)
			}
		}
	}

	if !m.points[row][col].opened {
		m.points[row][col].opened = true
		m.remain--
	}
}

func (m *model) setNumView() {
	m.num[0] = " 　 "
	m.num[1] = fmt.Sprintf("\x1b[36m%s\x1b[0m", " １ ")
	m.num[2] = fmt.Sprintf("\x1b[32m%s\x1b[0m", " ２ ")
	m.num[3] = fmt.Sprintf("\x1b[31m%s\x1b[0m", " ３ ")
	m.num[4] = fmt.Sprintf("\x1b[34m%s\x1b[0m", " ４ ")
	m.num[5] = fmt.Sprintf("\x1b[35m%s\x1b[0m", " ５ ")
	m.num[6] = fmt.Sprintf("\x1b[37m%s\x1b[0m", " ６ ")
	m.num[7] = fmt.Sprintf("\x1b[37m%s\x1b[0m", " ７ ")
	m.num[8] = fmt.Sprintf("\x1b[37m%s\x1b[0m", " ８ ")
}

func (m model) cell(x, y int) (int, int) {
	col := (x - marginLeft) / cellWidth
	row := (y - marginTop) / cellHeight

	return toDisplayNum(col, m.column), toDisplayNum(row, m.row)
}

func (m model) cellMiddle(data int, opened, flagged bool) string {

	d := "████"
	if opened {
		if data >= 0 {
			d = m.num[data]
		} else {
			d = bombIcon
		}
	} else if flagged {
		d = flagIcon
	}

	return d + vertical
}

func (m model) choose(col, row int) (tea.Model, tea.Cmd) {

	if m.points[row][col].flagged {
		return m, nil
	}

	if m.points[row][col].data == -1 {
		m.gameover = true
	}

	m.points[row][col].opened = true
	m.remain--

	if m.points[row][col].data == 0 {
		m.multiOpen(row, col)
	}

	return m, nil
}

func (m model) viewHeader() string {
	return head(m.bombnum)
}

func bottom(column int) string {
	return build("└", "┴", "┘", column)
}

func build(left, middle, right string, repeat int) string {
	border := left + horizontal + strings.Repeat(horizontal+middle+horizontal, repeat)
	border += horizontal + right + "\n"
	return withMarginLeft(border)
}

func cellRight() string {
	return "\n"
}

func cellLeft() string {
	return withMarginLeft(vertical)
}

func head(rem int) string {
	border := "┌" + horizontal + horizontal + "┐" + "\n"
	rems := fmt.Sprintf("\x1b[31m%s\x1b[0m", fmt.Sprintf(" %03d", rem))
	vrems := vertical + rems + vertical + "\n"
	border += withMarginLeft(vrems)
	return withMarginLeft(border)
}

func middle(column int) string {
	return build("├", "┼", "┤", column)
}

func toDisplayNum(num, max int) int {
	if num < 0 {
		num = 0
	} else if num > (max - 1) {
		num = -1
	}

	return num
}
func top(column int) string {
	left := "├" + horizontal + horizontal + "┼"
	return build(left, "┬", "┐", column-1)
}

func withMarginLeft(s string) string {
	return strings.Repeat(" ", marginLeft) + s
}

func gameover() string {
	return gameoverView
}

func gameclear() string {
	return gameclearView
}
