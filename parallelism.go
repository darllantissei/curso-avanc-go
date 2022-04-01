package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/eucatur/go-toolbox/database"
	"github.com/eucatur/go-toolbox/log"
)

type Parameter struct {
	ID          int64       `json:"id" db:"param_id"`
	Name        string      `json:"name" db:"param_name"`
	Value       string      `json:"value" db:"param_value"`
	Description string      `json:"description" db:"param_description"`
	Env         Environment `json:"environment" db:"param_environment"`
}

var queryNoRows = sql.ErrNoRows

var concurrencyDB sync.Mutex

const (
	configMySQLPROD    = "mysql-slave-prod.json"
	configMySQLSANDBOX = "mysql-slave-sandbox.json"
	queryParam         = "query_param.sql"
)

func checkParametersInProductionAndSandboxWithParallelism() {

	start := time.Now()

	parametersPROD := getListParam(environment_Production)

	parametersSANDBOX := getListParam(environment_Sandbox)

	fmt.Println("Iniciando o processo de verificação dos parâmetros")

	listParameterProd := <-parametersPROD

	listParameterSandbox := <-parametersSANDBOX

	parametersPendingInProduction := checkParameters(listParameterProd, listParameterSandbox)

	fmt.Println("Iniciando a preparação do script para inserção do parâmetros pendentes")

	generateContentWithParamsPending(parametersPendingInProduction)

	fmt.Println("\n\nQuantidade de parâmetros em produção: ", len(listParameterProd))
	fmt.Println("\n\nQuantidade de parâmetros em sandbox: ", len(listParameterSandbox))

	fmt.Println("\n\n\n\nTempo para execução: ", time.Since(start))

}

func generateContentWithParamsPending(parametersPending []Parameter) {

	if len(parametersPending) <= 0 {

		fmt.Println("Não há nenhum parâmetro pendente entre os ambientes")

		return
	}

	script := ""

	content, err := ioutil.ReadFile("insert_param.sql")

	if err != nil {
		panic(fmt.Errorf("ocorreu o seginte problema ao ler o templeta com a instrução INSERT dos parâmetros. Detalhes: %s", err.Error()))
	}

	insertData := string(content)

	for _, param := range parametersPending {

		script += fmt.Sprintln(fmt.Sprintf(insertData, param.Name, param.Value, param.Description))

	}

	fileScript := prepareFile()

	ioutil.WriteFile(fileScript, []byte(script), 0777)
}

func prepareFile() string {
	fileNameScriptParametersPending := "parameters.sql"

	_, errFileExits := os.Stat(fileNameScriptParametersPending)

	if errFileExits == nil {

		errFileExits = os.Remove(fileNameScriptParametersPending)

		if errFileExits != nil {
			panic(errFileExits)
		}
	}

	return fileNameScriptParametersPending
}

func checkParameters(parametersProd, parametersSandbox []Parameter) []Parameter {

	var wgp sync.WaitGroup
	var mtx sync.RWMutex

	paramtersPendingInProduction := []Parameter{}

	for _, paramSandbox := range parametersSandbox {

		wgp.Add(1)

		fmt.Println("Verificando o parâmetro: ", paramSandbox.Name)

		paramPendingInProduction := checkParamInProduction(paramSandbox, parametersProd, &wgp, &mtx)

		paramPending, hasValue := <-paramPendingInProduction
		if hasValue {
			paramtersPendingInProduction = append(paramtersPendingInProduction, paramPending)
		}

	}

	wgp.Wait()

	return paramtersPendingInProduction
}

func getListParam(typeEnv Environment) chan []Parameter {

	returnParameters := make(chan []Parameter)

	go func(env Environment) {

		fmt.Println("Consultando os parâmetro em ambiente de ", env.String())

		parameters, err := listParam(env)

		if err != nil {
			panic(fmt.Errorf("não foi possível listar o parâmetro em %s. Detalhes: %s", env.String(), err.Error()))
		}

		if env == environment_Sandbox {
			// make mocks parameters to tests
			parameters = append(parameters, []Parameter{
				{
					ID:          -1,
					Name:        "ParamTest",
					Value:       "ValueOfParamTest",
					Description: "Parâmetro falso para realização de teste",
					Env:         env,
				},
				{
					ID:          -2,
					Name:        "AnotherParamTest",
					Value:       "ValueOfParamTest",
					Description: "Outro parâmetro para realização de testes",
					Env:         env,
				},
				{
					ID:          -3,
					Name:        "MoreOneParamTest",
					Value:       "ValueOgMoreOneParamTest",
					Description: "Mais um parâmetro para realização de testes",
					Env:         env,
				},
			}...)
		}

		returnParameters <- parameters

		close(returnParameters)

	}(typeEnv)

	return returnParameters
}

func checkParamInProduction(paramToVerify Parameter, parametersToCheck []Parameter, wait *sync.WaitGroup, mx *sync.RWMutex) chan Parameter {

	paramPending := make(chan Parameter)

	go func(paramCheck Parameter, paramsToCheck []Parameter, wtg *sync.WaitGroup, mtx *sync.RWMutex) {

		defer wtg.Done()

		paramFound := false

		mtx.RLock()

		for _, paramProd := range parametersToCheck {

			fmt.Println("Checando parâmetros em produção ", paramProd.Name)

			if strings.EqualFold(paramProd.Name, paramCheck.Name) {
				fmt.Println("parâmetro ", paramProd.Name, " encontrado")
				paramFound = true
				break
			}
		}

		mtx.RUnlock()

		mtx.Lock()

		if !paramFound {

			fmt.Println("Parâmetro pendente em produção ", paramCheck.Name)

			paramPending <- paramCheck

		} else {

			fmt.Println("Parâmetro ", paramCheck.Name, " já existente no servidor")

		}

		close(paramPending)

		mtx.Unlock()

	}(paramToVerify, parametersToCheck, wait, mx)

	return paramPending

}

func listParam(env Environment) (parameters []Parameter, err error) {
	concurrencyDB.Lock()

	query, err := ioutil.ReadFile(queryParam)

	if err != nil {
		return
	}

	configDB := ""
	switch env {
	case environment_Sandbox:
		configDB = configMySQLSANDBOX
	case environment_Production:
		configDB = configMySQLPROD
	}

	dml := string(query)

	err = database.MustGetByFile(configDB).Select(&parameters, dml)

	if err != nil && err != queryNoRows {
		log.Error(err)

		err = fmt.Errorf("Não foi possível listar os parâmetros para o ambiente %s ", env.String())
	}

	concurrencyDB.Unlock()

	return
}
