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
  storage_encrypted       = true
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

# Replication group (not a bare cluster) so encryption at rest + in transit +
# an AUTH token are available. Single node — the cache is rebuildable from Mongo.
resource "aws_elasticache_replication_group" "redis" {
  replication_group_id       = "${var.project}-redis"
  description                = "flagcast flag cache"
  engine                     = "redis"
  node_type                  = "cache.t3.micro"
  num_cache_clusters         = 1
  parameter_group_name       = "default.redis7"
  port                       = 6379
  subnet_group_name          = aws_elasticache_subnet_group.main.name
  security_group_ids         = [aws_security_group.data.id]
  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
  auth_token                 = var.redis_auth_token
  automatic_failover_enabled = false
}
