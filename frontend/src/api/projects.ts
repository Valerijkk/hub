import { api } from "./client";

// --- Request types ---

interface CreateProjectRequest {
  name: string;
  key: string;
  description?: string;
}

interface UpdateProjectRequest {
  name?: string;
  description?: string;
  avatar_url?: string;
}

interface AddMemberRequest {
  user_id: string;
  role: "developer" | "tester" | "viewer";
}

interface CreateInviteRequest {
  role: "developer" | "tester" | "viewer";
  max_uses?: number;
  expires_at?: string;
}

// --- Response types ---

export interface ProjectResponse {
  id: string;
  name: string;
  key: string;
  description?: string;
  avatar_url?: string;
  status: string;
  member_count: number;
  my_role?: string;
  created_at: string;
  updated_at: string;
}

export interface MemberResponse {
  user_id: string;
  email: string;
  name: string;
  avatar_url?: string;
  role: string;
  joined_at: string;
}

export interface InviteResponse {
  id: string;
  code: string;
  role: string;
  max_uses: number;
  use_count: number;
  expires_at?: string;
  link: string;
  created_at: string;
}

// --- API functions ---

export function listProjects() {
  return api.get<ProjectResponse[]>("/projects");
}

export function createProject(data: CreateProjectRequest) {
  return api.post<ProjectResponse>("/projects", data);
}

export function getProject(projectId: string) {
  return api.get<ProjectResponse>(`/projects/${projectId}`);
}

export function updateProject(projectId: string, data: UpdateProjectRequest) {
  return api.put<ProjectResponse>(`/projects/${projectId}`, data);
}

export function archiveProject(projectId: string) {
  return api.post<void>(`/projects/${projectId}/archive`);
}

export function deleteProject(projectId: string) {
  return api.delete<void>(`/projects/${projectId}`);
}

export function restoreProject(projectId: string) {
  return api.post<void>(`/projects/${projectId}/restore`);
}

export function listMembers(projectId: string) {
  return api.get<MemberResponse[]>(`/projects/${projectId}/members`);
}

export function addMember(projectId: string, data: AddMemberRequest) {
  return api.post<MemberResponse>(`/projects/${projectId}/members`, data);
}

export function removeMember(projectId: string, userId: string) {
  return api.delete<void>(`/projects/${projectId}/members/${userId}`);
}

export function createInvite(projectId: string, data: CreateInviteRequest) {
  return api.post<InviteResponse>(`/projects/${projectId}/invites`, data);
}

export function joinByInvite(code: string) {
  return api.post<void>("/projects/join", { code });
}
