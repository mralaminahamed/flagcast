resource "aws_ecs_cluster" "main" {
  name = var.project
  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

# Internal DNS so services resolve each other as <name>.flagcast.local.
resource "aws_service_discovery_private_dns_namespace" "main" {
  name = "flagcast.local"
  vpc  = aws_vpc.main.id
}

locals {
  mongo_uri = "mongodb://flagcast:${var.docdb_password}@${aws_docdb_cluster.main.endpoint}:27017/flagcast?tls=true&retryWrites=false"
  redis_url = "redis://${aws_elasticache_cluster.redis.cache_nodes[0].address}:6379/0"
  nats_url  = "nats://nats.flagcast.local:4222"

  # One entry per Fargate service. `lb = true` puts it behind the ALB.
  services = {
    nats = {
      image       = "nats:2-alpine"
      port        = 4222
      cpu         = 256
      memory      = 512
      command     = ["-js", "-m", "8222"]
      lb          = false
      health_path = ""
      extra_ports = []
      env         = {}
    }
    gateway = {
      image       = "${var.image_registry}/gateway:${var.image_tag}"
      port        = 8080
      cpu         = 256
      memory      = 512
      command     = []
      lb          = true
      health_path = "/health"
      extra_ports = []
      env = {
        PORT            = "8080"
        MONGO_URI       = local.mongo_uri
        REDIS_URL       = local.redis_url
        NATS_URL        = local.nats_url
        AI_URL          = "http://ai.flagcast.local:8090"
        GATEWAY_API_KEY = var.gateway_api_key
      }
    }
    evaluator = {
      image       = "${var.image_registry}/evaluator:${var.image_tag}"
      port        = 8081
      cpu         = 256
      memory      = 512
      command     = []
      lb          = false
      health_path = "/health"
      extra_ports = [50051]
      env = {
        PORT      = "8081"
        GRPC_ADDR = ":50051"
        MONGO_URI = local.mongo_uri
        REDIS_URL = local.redis_url
        NATS_URL  = local.nats_url
      }
    }
    ai = {
      image       = "${var.image_registry}/ai:${var.image_tag}"
      port        = 8090
      cpu         = 256
      memory      = 512
      command     = []
      lb          = false
      health_path = "/health"
      extra_ports = []
      env = {
        PORT              = "8090"
        MONGO_URI         = local.mongo_uri
        ANTHROPIC_API_KEY = var.anthropic_api_key
      }
    }
  }
}

resource "aws_cloudwatch_log_group" "svc" {
  for_each          = local.services
  name              = "/ecs/${var.project}/${each.key}"
  retention_in_days = 14
}

resource "aws_service_discovery_service" "svc" {
  for_each = local.services
  name     = each.key
  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.main.id
    dns_records {
      type = "A"
      ttl  = 10
    }
    routing_policy = "MULTIVALUE"
  }
  health_check_custom_config {
    failure_threshold = 1
  }
}

resource "aws_ecs_task_definition" "svc" {
  for_each                 = local.services
  family                   = "${var.project}-${each.key}"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = each.value.cpu
  memory                   = each.value.memory
  execution_role_arn       = aws_iam_role.task_exec.arn

  container_definitions = jsonencode([
    {
      name      = each.key
      image     = each.value.image
      essential = true
      command   = each.value.command
      portMappings = [
        for p in concat([each.value.port], each.value.extra_ports) :
        { containerPort = p, protocol = "tcp" }
      ]
      environment = [for k, v in each.value.env : { name = k, value = v }]
      # ECS-level container health check (services expose /health; nats has none).
      healthCheck = each.value.health_path == "" ? null : {
        command     = ["CMD-SHELL", "wget -qO- http://localhost:${each.value.port}${each.value.health_path} || exit 1"]
        interval    = 30
        timeout     = 5
        retries     = 3
        startPeriod = 15
      }
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.svc[each.key].name
          "awslogs-region"        = var.region
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])
}

resource "aws_ecs_service" "svc" {
  for_each        = local.services
  name            = each.key
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.svc[each.key].arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.public[*].id
    security_groups  = [aws_security_group.service.id]
    assign_public_ip = true
  }

  # Auto-rollback a deploy that fails to reach steady state (bad image, crashloop).
  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }

  service_registries {
    registry_arn = aws_service_discovery_service.svc[each.key].arn
  }

  dynamic "load_balancer" {
    for_each = each.value.lb ? [1] : []
    content {
      target_group_arn = aws_lb_target_group.gateway.arn
      container_name   = each.key
      container_port   = each.value.port
    }
  }

  depends_on = [aws_lb_listener.http]
}
