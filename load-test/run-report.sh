#!/usr/bin/env bash

generateResults() {
    echo " " > TEMP-OK.md
    echo "| participante | multa SLA (> 249ms) | multa SLA (inconsistência saldo) | multa total | valor a receber | relatório |" >> TEMP-OK.md
    echo "| --           | --                  | --                               | --          | --              | --        |" >> TEMP-OK.md

    valorContrato=100000.0
    SLARespostasOk=98.0
    multaInconsistenciaSaldoLimiteUnidade=803.01
   
   echo "computando participantes..."
    for diretorio in user-files/results/*/; do
    (
        reportFileCount=$(find $diretorio -name index.html | wc -l)

        if [ $reportFileCount -eq "1" ]; then

            arquivoStats=$(find $diretorio -name stats.json)
            reportFile=$(find $diretorio -name index.html)
            simulationFile=$(find $diretorio -name simulation.log)
            reportDir=$(dirname $reportFile)

            totalRequests=$(cat $arquivoStats | jq '.stats.numberOfRequests.total')
            responsesOkMenos250ms=$(cat $arquivoStats | jq '.stats.group1.count')
            porcentagemRespostasAceitaveis=$(python3 -c "print(round(${responsesOkMenos250ms} / ${totalRequests} * 100, 2))")
            inconsistenciasSaldoLimite=$(grep "ConsistenciaSaldoLimite" $simulationFile | wc -l)
            inconsistenciaTransacoesSaldo=$(grep "jmesPath(saldo.total).find.is" $simulationFile | wc -l)
            multaSLA250ms=$(python3 -c "print(max(0.0, round(((${SLARespostasOk} - ${porcentagemRespostasAceitaveis}) * 1000), 2)))")
            multaSLAInconsSaldo=$(python3 -c "print(round(((${inconsistenciasSaldoLimite} + ${inconsistenciaTransacoesSaldo}) * ${multaInconsistenciaSaldoLimiteUnidade}), 2))")
            multaSLATotal=$(python3 -c "print(round(${multaSLA250ms} + ${multaSLAInconsSaldo}, 2))")
            pagamento=$(python3 -c "print(max(0.0, round(${valorContrato} - ${multaSLATotal}, 2)))")

            echo -n "| David Alecrim " >> TEMP-OK.md
            echo -n "| USD ${multaSLA250ms} " >> TEMP-OK.md
            echo -n "| USD ${multaSLAInconsSaldo} " >> TEMP-OK.md
            echo -n "| USD ${multaSLATotal} " >> TEMP-OK.md
            echo -n "| **USD ${pagamento}** " >> TEMP-OK.md
            echo    "| [link]($reportDir) |" >> TEMP-OK.md
        fi
    )
    done

    user-files/cat RESULTADOS-HEADER.md > RESULTADOS.md
    cat TEMP-OK.md >> RESULTADOS.md

    echo " " >> RESULTADOS.md

    rm TEMP-OK.md
}

generateResults