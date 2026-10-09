package views

// import (
// 	"fmt"
// 	"os"

// 	"charm.land/bubbles/v2/list"
// 	tea "charm.land/bubbletea/v2"
// 	"charm.land/huh/v2"
// )

// // 1. O tipo de dado que queremos manipular
// type Usuario struct {
// 	Nome     string
// 	Telefone string
// 	Email    string
// }

// // Métodos necessários para que o Usuário seja renderizável na lista de "bubbles/list"
// func (u Usuario) Title() string       { return u.Nome }
// func (u Usuario) Description() string { return fmt.Sprintf("%s | %s", u.Telefone, u.Email) }
// func (u Usuario) FilterValue() string { return u.Nome }

// // 2. Os estados possíveis da nossa aplicação
// type sessionState int

// const (
// 	stateList sessionState = iota
// 	stateForm
// )

// // 3. O Model principal do Bubble Tea
// type MainModel struct {
// 	state     sessionState
// 	usuarios  []Usuario
// 	listaView list.Model
// 	formView  *huh.Form
// 	editIndex int // Guarda o índice se for edição, ou -1 se for novo usuário

// 	// Variáveis temporárias que o formulário do huh irá preencher via ponteiro
// 	tempNome     string
// 	tempTelefone string
// 	tempEmail    string
// }

// func NewMainModel() MainModel {
// 	// Dados iniciais de exemplo
// 	itensIniciais := []list.Item{
// 		Usuario{Nome: "João Silva", Telefone: "11 99999-9999", Email: "joao@email.com"},
// 		Usuario{Nome: "Maria Souza", Telefone: "21 98888-8888", Email: "maria@email.com"},
// 	}

// 	l := list.New(itensIniciais, list.NewDefaultDelegate(), 0, 0)
// 	l.Title = "Usuários Cadastrados ('a' para add, 'e' para editar)"

// 	return MainModel{
// 		state:     stateList,
// 		listaView: l,
// 		editIndex: -1,
// 	}
// }

// // 4. Função auxiliar para criar/reiniciar o formulário HUH dinamicamente
// func (m *MainModel) CarregarFormulario(u *Usuario) {
// 	if u != nil {
// 		// Se passou um usuário, estamos EDITANDO (carrega dados existentes)
// 		m.tempNome = u.Nome
// 		m.tempTelefone = u.Telefone
// 		m.tempEmail = u.Email
// 	} else {
// 		// Se for NIL, estamos ADICIONANDO (limpa os campos)
// 		m.tempNome = ""
// 		m.tempTelefone = ""
// 		m.tempEmail = ""
// 	}

// 	// Criamos um novo huh.Form apontando para os campos temporários do nosso modelo
// 	m.formView = huh.NewForm(
// 		huh.NewGroup(
// 			huh.NewInput().Title("Nome").Value(&m.tempNome),
// 			huh.NewInput().Title("Telefone").Value(&m.tempTelefone),
// 			huh.NewInput().Title("E-mail").Value(&m.tempEmail),
// 		),
// 	)

// 	// Inicializa o formulário (necessário no ciclo do bubbletea)
// 	m.formView.Init()
// }

// // --- Ciclo de Vida do Bubble Tea ---

// func (m MainModel) Init() tea.Cmd {
// 	return nil
// }

// func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	var cmd tea.Cmd
// 	var cmds []tea.Cmd

// 	switch m.state {
// 	case stateList:
// 		// --- LÓGICA DA LISTA ---
// 		switch msg := msg.(type) {
// 		case tea.KeyMsg:
// 			switch msg.String() {
// 			case "a": // Tecla 'a' para Adicionar
// 				m.editIndex = -1
// 				m.CarregarFormulario(nil) // Cria form limpo
// 				m.state = stateForm
// 				return m, nil

// 			case "e": // Tecla 'e' para Editar
// 				if idx := m.listaView.Index(); idx >= 0 && idx < len(m.listaView.Items()) {
// 					m.editIndex = idx
// 					usuarioSelecionado := m.listaView.SelectedItem().(Usuario)
// 					m.CarregarFormulario(&usuarioSelecionado) // Cria form com dados
// 					m.state = stateForm
// 					return m, nil
// 				}
// 			}
// 		}
// 		// Se não mudou de tela, repassa as mensagens padrão para atualizar a lista
// 		m.listaView, cmd = m.listaView.Update(msg)
// 		return m, cmd

// 	case stateForm:
// 		// --- LÓGICA DO FORMULÁRIO (HUH) ---

// 		// Atualiza o formulário interno repassando a mensagem do Bubble Tea
// 		formModel, formCmd := m.formView.Update(msg)
// 		if f, ok := formModel.(*huh.Form); ok {
// 			m.formView = f
// 			cmds = append(cmds, formCmd)
// 		}

// 		// O Huh altera seu estado interno para Completed quando o usuário envia o formulário
// 		if m.formView.State == huh.StateCompleted {
// 			novoUsuario := Usuario{
// 				Nome:     m.tempNome,
// 				Telefone: m.tempTelefone,
// 				Email:    m.tempEmail,
// 			}

// 			if m.editIndex == -1 {
// 				// Adicionando um novo registro
// 				m.listaView.InsertItem(len(m.listaView.Items()), novoUsuario)
// 			} else {
// 				// Atualizando um registro existente
// 				m.listaView.SetItem(m.editIndex, novoUsuario)
// 			}

// 			// Volta para o estado da lista
// 			m.state = stateList
// 		}

// 		// Permite cancelar e voltar se o formulário for abortado (ex: pressionando Esc)
// 		if m.formView.State == huh.StateAborted {
// 			m.state = stateList
// 		}

// 		return m, tea.Batch(cmds...)
// 	}

// 	return m, nil
// }

// func (m MainModel) View() string {
// 	switch m.state {
// 	case stateForm:
// 		return m.formView.View()
// 	default:
// 		return m.listaView.View()
// 	}
// }

// func main() {
// 	p := tea.NewProgram(NewMainModel())
// 	if _, err := p.Run(); err != nil {
// 		fmt.Printf("Erro na execução: %v", err)
// 		os.Exit(1)
// 	}
// }
