# Managed datastores: DocumentDB (Mongo-compatible) and ElastiCache Redis.

resource "aws_docdb_subnet_group" "main" {
  name       = "${var.project}-docdb"
  subnet_ids = aws_subnet.private[*].id
}

resource "aws_docdb_cluster" "main" {
  cluster_identifier      = "${var.project}-docdb"
  engine                  = "docdb"
  master_username         = "flagcast"
  master_password         = var.docdb_password
  db_subnet_group_name    = aws_docdb_subnet_group.main.name
  vpc_security_group_ids  = [aws_security_group.data.id]
  skip_final_snapshot     = true
  deletion_protection     = false
  backup_retention_period = 1
}

resource "aws_docdb_cluster_instance" "main" {
  identifier         = "${var.project}-docdb-0"
  cluster_identifier = aws_docdb_cluster.main.id
  instance_class     = "db.t3.medium"
}

resource "aws_elasticache_subnet_group" "main" {
  name       = "${var.project}-redis"
  subnet_ids = aws_subnet.private[*].id
}

resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "${var.project}-redis"
  engine               = "redis"
  node_type            = "cache.t3.micro"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis7"
  port                 = 6379
  subnet_group_name    = aws_elasticache_subnet_group.main.name
  security_group_ids   = [aws_security_group.data.id]
}
