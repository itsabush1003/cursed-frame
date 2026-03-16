import { memo } from "react";

import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from "@tanstack/react-table";

export interface rowData {
  userName: string;
  teamId: number;
  isReady: boolean;
  reject: () => void;
}

const UsersTable = memo(({ data }: { data: rowData[] }) => {
  const columnHelper = createColumnHelper<rowData>();
  const columns = [
    columnHelper.accessor("userName", { header: "名前" }),
    columnHelper.accessor("isReady", { header: "準備完了" }),
    columnHelper.accessor("teamId", { header: "チーム" }),
    columnHelper.display({
      header: "参加を取り消す",
      id: "reject",
      cell: (prop) => (
        <button onClick={() => prop.row.original.reject()}>取り消す</button>
      ),
    }),
  ];

  const table = useReactTable<rowData>({
    data: data,
    columns: columns,
    getCoreRowModel: getCoreRowModel(),
  });

  return (
    <>
      <table>
        <thead>
          {table.getHeaderGroups().map((headerGroup) => (
            <tr key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <th>
                  {header.isPlaceholder
                    ? null
                    : flexRender(
                        header.column.columnDef.header,
                        header.getContext(),
                      )}
                </th>
              ))}
            </tr>
          ))}
        </thead>
        <tbody>
          {table.getRowModel().rows.map((row) => (
            <tr key={row.id}>
              {row.getVisibleCells().map((cell) => (
                <td key={cell.id}>
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </>
  );
});

export default UsersTable;
