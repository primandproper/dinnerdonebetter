import { reportStatusToJSON, type ScopedIssueReport } from '@primandproper/platform-client/issuereports/v1';

/**
 * OperatorRow is one report in the operator's queue, flattened for the table. The account is
 * carried beside the report because this queue spans every account, and a row that could not
 * say whose report it was could not be acted on.
 */
export interface OperatorRow {
  id: string;
  accountID: string;
  kind: string;
  status: string;
  subject: string;
  createdAt: Date | undefined;
}

export function toOperatorRows(results: ScopedIssueReport[]): OperatorRow[] {
  const rows: OperatorRow[] = [];

  for (const scoped of results) {
    const report = scoped.report;
    if (!report) {
      continue;
    }

    rows.push({
      id: report.id,
      accountID: scoped.scope,
      kind: report.kind,
      status: reportStatusToJSON(report.status),
      subject: report.subjectType ? `${report.subjectType}/${report.subjectId}` : '',
      createdAt: report.createdAt,
    });
  }

  return rows;
}
