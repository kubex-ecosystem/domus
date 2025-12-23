package infra

import (
	"fmt"

	ci "github.com/kubex-ecosystem/domus/internal/interfaces"
	"github.com/kubex-ecosystem/domus/internal/module/kbx"

	// kbxKrs "github.com/kubex-ecosystem/kbx/tools/security/external"

	"github.com/docker/go-connections/nat"
	svc "github.com/kubex-ecosystem/domus/internal/services/docker"
	"github.com/kubex-ecosystem/domus/internal/types"
	logz "github.com/kubex-ecosystem/logz"
)

func SetupRabbitMQ(config *types.DBConfig, dockerService ci.IDockerService) error {
	if config == nil || !kbx.DefaultTrue(config.Enabled) {
		logz.Log("debug", "RabbitMQ está desabilitado na configuração. Ignorando inicialização.")
		return nil
	}

	// Verifica se o serviço já está rodando
	if svc.IsServiceRunning(config.Name) {
		logz.Successf("RabbitMQ (%s) já está rodando!", config.Name)
		return nil
	}

	if config.User == "" {
		config.User = "gobe"
	}
	// if config.Pass == "" {
	// TODO: Mudar para usar o serviço de fileKeyring do kbx, REMOVER COMPLETAMENTE keyring (dbus) do projeto
	// rabbitPassKey, rabbitPassErr := kbxKrs.GetOrGenerateFromKeyring("rabbitmq")
	// if rabbitPassErr != nil {
	// 	logz.Log("error", "Skipping RabbitMQ setup due to error generating password")
	// 	logz.Log("debug", fmt.Sprintf("Error generating key: %v", rabbitPassErr))
	// 	return rabbitPassErr
	// } else {
	// 	config.Pass = string(rabbitPassKey)
	// }
	// }

	// if config.Vhost == "" {
	// 	config.Vhost = "gobe"
	// }
	// if config.Port == "" {
	// 	config.Port = "5672"
	// }
	// if config.ManagementPort == "" {
	// 	config.ManagementPort = "15672"
	// }
	// if config.ErlangCookie == "" {
	// 	rabbitCookieKey, rabbitCookieErr := krg.GetOrGenerateFromKeyring("rabbitmq-cookie")
	// 	if rabbitCookieErr != nil {
	// 		logz.Log("error", "Skipping RabbitMQ setup due to error generating password")
	// 		logz.Log("debug", fmt.Sprintf("Error generating key: %v", rabbitCookieErr))
	// 		return rabbitCookieErr
	// 	} else {
	// 		config.ErlangCookie = string(rabbitCookieKey)
	// 	}
	// }
	// if config.Volume == "" {
	// 	config.Volume = os.ExpandEnv(kbx.DefaultRabbitMQVolume)
	// }
	// // Cria o volume, se necessário
	// if err := dockerService.CreateVolume(config.Name, config.Volume); err != nil {
	// 	logz.Log("error", fmt.Sprintf("❌ Erro ao criar volume do RabbitMQ: %v", err))
	// 	return err
	// }

	// Mapeia as portas

	portBindings := []nat.PortMap{
		{
			nat.Port(fmt.Sprintf("%s/tcp", config.Port)): []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: fmt.Sprintf("%v", config.Port)}}, // publica AMQP
			// nat.Port(fmt.Sprintf("%s/tcp", config.ManagementPort)): []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: fmt.Sprintf("%v", config.ManagementPort)}}, // publica console
		},
	}

	// Configura as variáveis de ambiente
	envVars := map[string]string{
		"RABBITMQ_DEFAULT_USER": config.User,
		"RABBITMQ_DEFAULT_PASS": config.Pass,
		// "RABBITMQ_DEFAULT_VHOST=" + config.Vhost,
		// "RABBITMQ_ERLANG_COOKIE=" + config.ErlangCookie,
	}

	// Inicializa o container do RabbitMQ
	service := dockerService.AddService(
		config.Name,
		"rabbitmq:latest",
		envVars,
		portBindings,
		map[string]struct{}{
			// fmt.Sprintf("%s:/var/lib/rabbitmq", config.Volume): {},
		},
	)
	if service == nil {
		err := fmt.Errorf("serviço não encontrado: %s", config.Name)
		logz.Errorf("Erro ao iniciar o RabbitMQ: %v", err)
		return err
	}

	logz.Successf("RabbitMQ (%s) iniciado com sucesso!", config.Name)
	return nil
}
