[comment]: <> (Thanks to k3s and the original version of this document is 
https://rancher.com/docs/k3s/latest/en/installation/datastore/#datastore-endpoint-format-and-functionality)

# Database endpoint format

As mentioned in readme, the format of value passed to the datastore-endpoint parameter is dependent upon the datastore 
backend. The following details this format and functionality for each supported external datastore.

## PostgreSQL

In its most common form, the datastore-endpoint parameter for PostgreSQL has the following format:

`postgres://username:password@hostname:port/database-name`

More advanced configuration parameters are available. For more information on these, please see 
https://godoc.org/github.com/lib/pq.

If you specify a database name and it does not exist, the server will attempt to create it.

If you only supply `postgres://` as the endpoint, velad will attempt to do the following:

- Connect to localhost using `postgres` as the username and password
- Create a database named `kubernetes`

## MySQL/MariaDB

In its most common form, the datastore-endpoint parameter for MySQL and MariaDB has the following format:

`mysql://username:password@tcp(hostname:3306)/database-name`

More advanced configuration parameters are available. For more information on these, please see 
https://github.com/go-sql-driver/mysql#dsn-data-source-name

Note that due to a [known issue](https://github.com/rancher/k3s/issues/1093) in K3s, you cannot set the `tls` parameter. TLS communication is supported, but you cannot, for example, set this parameter to “skip-verify” to cause K3s to skip certificate verification.

If you specify a database name and it does not exist, the server will attempt to create it.

If you only supply `mysql://` as the endpoint, K3s will attempt to do the following:

- Connect to the MySQL socket at `/var/run/mysqld/mysqld.sock` using the `root` user and no password
- Create a database with the name `kubernetes`

## etcd

In its most common form, the datastore-endpoint parameter for etcd has the following format:

`https://etcd-host-1:2379,https://etcd-host-2:2379,https://etcd-host-3:2379`

The above assumes a typical three node etcd cluster. The parameter can accept one more comma separated etcd URLs.