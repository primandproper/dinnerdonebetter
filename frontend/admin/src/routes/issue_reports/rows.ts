import { reportStatusToJSON, type IssueReport } from '@primandproper/platform-client/issuereports/v1';

/**
 * OperatorRow is one report in the service's queue, flattened for the table. The reporter is
 * carried beside the report because every report in the deployment is in this one queue, and a
 * row that could not say whose report it was could not be acted on.
 */
export interface OperatorRow {
  id: string;
  reporter: string;
  kind: string;
  status: string;
  subject: string;
  createdAt: Date | undefined;
}

export function toOperatorRows(results: IssueReport[]): OperatorRow[] {
  return results.map((report) => ({
    id: report.id,
    reporter: report.reporter,
    kind: report.kind,
    status: reportStatusToJSON(report.status),
    subject: report.subjectType ? `${report.subjectType}/${report.subjectId}` : '',
    createdAt: report.createdAt,
  }));
}
