variable "mysql_db_name" { type = string }
variable "mysql_hostname" { type = string }
variable "port" {
    type = number
    default = 3306 
}
variable "admin_username" { type = string }
variable "admin_password" { type = string }
variable "use_tls" { type = bool }
