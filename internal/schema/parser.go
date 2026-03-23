package schema

import (
	// "fmt"
	pg_query "github.com/pganalyze/pg_query_go/v5"
)
func parseDDLStatement(sql string) (*pg_query.ParseResult, error) {
	astTree, err := pg_query.Parse(sql)
	if err != nil {
		return nil, err
	}
	return astTree, nil
}

// CREATE TABLE users (
// 		id INT PRIMARY KEY,
// 		email TEXT
// 	);

// output of ast parser
// {
//   "version": 160001,
//   "stmts": [
//     {
//       "stmt": {
//         "CreateStmt": {
//           "relation": {
//             "relname": "users",
//             "inh": true,
//             "relpersistence": "p",
//             "location": 15
//           },
//           "tableElts": [
//             {
//               "ColumnDef": {
//                 "colname": "id",
//                 "typeName": {
//                   "names": [
//                     {
//                       "String": {
//                         "sval": "pg_catalog"
//                       }
//                     },
//                     {
//                       "String": {
//                         "sval": "int4"
//                       }
//                     }
//                   ],
//                   "typemod": -1,
//                   "location": 28
//                 },
//                 "is_local": true,
//                 "constraints": [
//                   {
//                     "Constraint": {
//                       "contype": "CONSTR_PRIMARY",
//                       "location": 32
//                     }
//                   }
//                 ],
//                 "location": 25
//               }
//             },
//             {
//               "ColumnDef": {
//                 "colname": "email",
//                 "typeName": {
//                   "names": [
//                     {
//                       "String": {
//                         "sval": "text"
//                       }
//                     }
//                   ],
//                   "typemod": -1,
//                   "location": 53
//                 },
//                 "is_local": true,
//                 "location": 47
//               }
//             }
//           ],
//           "oncommit": "ONCOMMIT_NOOP"
//         }
//       },
//       "stmt_len": 60
//     }
//   ]
// }
